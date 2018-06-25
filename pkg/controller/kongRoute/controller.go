package kongRoute

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"time"
	// 	kanaryv1 "github.com/etiennecoutaud/kanary/pkg/apis/kanary/v1"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	// 	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// 	"k8s.io/apimachinery/pkg/runtime/schema"
	apimv1alpha1 "github.com/cdiscount/kong-operator/pkg/apis/apim/v1alpha1"
	clientset "github.com/cdiscount/kong-operator/pkg/client/clientset/versioned"
	apimscheme "github.com/cdiscount/kong-operator/pkg/client/clientset/versioned/scheme"
	informers "github.com/cdiscount/kong-operator/pkg/client/informers/externalversions"
	listers "github.com/cdiscount/kong-operator/pkg/client/listers/apim/v1alpha1"
	utils "github.com/cdiscount/kong-operator/pkg/utils"
	kongClient "github.com/etiennecoutaud/kong-client-go/kong"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	maxRetries = 15
	// controllerAgentName is the name in event sources
	controllerAgentName = "kongRoute-controller"

	// SuccessSynced is used as part of the Event 'reason' when a Database is synced
	SuccessSynced = "Synced"
)

// Controller is the controller implementation for kanary resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	apimClientset clientset.Interface

	kongRouteLister   listers.KongRouteLister
	kongServiceLister listers.KongServiceLister

	kongRouteListerSynced   cache.InformerSynced
	kongServiceListerSynced cache.InformerSynced
	// apim that needto be synced
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder   record.EventRecorder
	kongClient *kongClient.Client
}

// NewController returns a new kanary controller
func NewController(
	kubeclientset kubernetes.Interface,
	apimClientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	apimInformerFactory informers.SharedInformerFactory,
	kongClient *kongClient.Client) *Controller {

	kongRouteInformer := apimInformerFactory.Apim().V1alpha1().KongRoutes()
	kongServiceInformer := apimInformerFactory.Apim().V1alpha1().KongServices()
	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	apimscheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:           kubeclientset,
		apimClientset:           apimClientset,
		kongRouteLister:         kongRouteInformer.Lister(),
		kongRouteListerSynced:   kongRouteInformer.Informer().HasSynced,
		kongServiceLister:       kongServiceInformer.Lister(),
		kongServiceListerSynced: kongServiceInformer.Informer().HasSynced,
		workqueue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kongRoute"),
		recorder:                recorder,
		kongClient:              kongClient,
	}

	kongRouteInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addKongRouteHandler,
		UpdateFunc: controller.updateKongRouteHandler,
		DeleteFunc: controller.deleteKongRouteHandler,
	})

	glog.Info("Setting up event handlers")
	return controller
}

func (c *Controller) addKongRouteHandler(obj interface{}) {
	kongRoute := obj.(*apimv1alpha1.KongRoute)
	glog.Info("Add KongRoute ", kongRoute.Name)
	c.enqueue(kongRoute)
}

func (c *Controller) updateKongRouteHandler(old interface{}, cur interface{}) {
	oldKongRoute := old.(*apimv1alpha1.KongRoute)
	kongRoute := cur.(*apimv1alpha1.KongRoute)
	if !(reflect.DeepEqual(oldKongRoute.Spec, kongRoute.Spec)) {
		glog.Info("Update KongRoute:", kongRoute.Name)
		c.enqueue(kongRoute)
	}
}

func (c *Controller) deleteKongRouteHandler(obj interface{}) {
	kongRoute := obj.(*apimv1alpha1.KongRoute)
	if !reflect.DeepEqual(kongRoute.Status, apimv1alpha1.KongRouteStatus{}) {
		err := c.deleteRoute(kongRoute)
		if err != nil {
			glog.Fatalf("Error when delete KongService: %s => %s", kongRoute.Name, err)
		}
		glog.Infof("KongService: %s deleted", kongRoute.Name)
	}
}

// enqueue takes a KongRoute resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Db.
func (c *Controller) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting KongRoute controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.kongRouteListerSynced, c.kongServiceListerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process Database resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (c *Controller) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	key, quit := c.workqueue.Get()
	if quit {
		return false
	}
	defer c.workqueue.Done(key)
	err := c.reconcile(key.(string))
	c.handleErr(err, key)

	return true
}

func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		c.workqueue.Forget(key)
		return
	}

	if c.workqueue.NumRequeues(key) < maxRetries {
		glog.V(2).Infof("Error syncing kongRoute %v: %v", key, err)
		c.workqueue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	glog.V(2).Infof("Dropping kongRoute %q out of the queue: %v", key, err)
	c.workqueue.Forget(key)
}

// Sync loop for kongRoute resources
func (c *Controller) reconcile(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the KongRoute resource with this namespace/name
	kgr, err := c.kongRouteLister.KongRoutes(namespace).Get(name)
	if err != nil {
		// processing.
		if errors.IsNotFound(err) {
			glog.V(2).Info(fmt.Sprintf("KongRoute %s no longer exists", key))
			return nil
		}
		return err
	}

	// Create new KongRoute
	// If no status then new object
	// Else need to update
	if reflect.DeepEqual(kgr.Status, apimv1alpha1.KongRouteStatus{}) {
		kgr, err = c.createRoute(kgr)
		if err != nil {
			c.recorder.Event(kgr, corev1.EventTypeNormal, "Error", err.Error())
			return err
		}
		c.recorder.Event(kgr, corev1.EventTypeNormal, "Info", "Kong Route created")
	} else {
		kgr, err = c.updateRoute(kgr)
		if err != nil {
			c.recorder.Event(kgr, corev1.EventTypeNormal, "Error", err.Error())
			return err
		}
		c.recorder.Event(kgr, corev1.EventTypeNormal, "Info", "Kong Route updated")
	}

	c.recorder.Event(kgr, corev1.EventTypeNormal, "Sync", "Correctly sync")
	return nil
}

func (c *Controller) deleteRoute(kgr *apimv1alpha1.KongRoute) error {
	_, err := c.kongClient.Route.Delete(kgr.Status.KongID)
	return err
}

func (c *Controller) updateRoute(kgr *apimv1alpha1.KongRoute) (*apimv1alpha1.KongRoute, error) {

	serviceRef, err := c.kongServiceLister.KongServices(kgr.Namespace).Get(kgr.Spec.ServiceName)
	if err != nil {
		return kgr, err
	}

	kongRouteAPI := &kongClient.Route{
		Protocols:    kgr.Spec.Protocols,
		Methods:      kgr.Spec.Methods,
		Hosts:        kgr.Spec.Hosts,
		Paths:        kgr.Spec.Paths,
		StripPath:    kgr.Spec.StripPath,
		PreserveHost: kgr.Spec.PreserveHost,
		Service: &kongClient.ServiceRef{
			ID: serviceRef.Status.KongID,
		},
	}

	ret, err := c.kongClient.Route.Patch(kgr.Status.KongID, kongRouteAPI)
	if err != nil {
		glog.V(4).Infof("Return from kong: Error Code=%d, %v", ret.StatusCode, ret.Body)
		return kgr, err
	}
	kgcr, err := unmarshalRoute(ret.Body)
	if err != nil {
		return kgr, err
	}
	return c.updateStatus(kgr, kgcr)
}

func (c *Controller) createRoute(kgr *apimv1alpha1.KongRoute) (*apimv1alpha1.KongRoute, error) {

	serviceRef, err := c.kongServiceLister.KongServices(kgr.Namespace).Get(kgr.Spec.ServiceName)
	if err != nil {
		return kgr, err
	}

	kongRouteAPI := &kongClient.Route{
		Protocols:    kgr.Spec.Protocols,
		Methods:      kgr.Spec.Methods,
		Hosts:        kgr.Spec.Hosts,
		Paths:        kgr.Spec.Paths,
		StripPath:    kgr.Spec.StripPath,
		PreserveHost: kgr.Spec.PreserveHost,
		Service: &kongClient.ServiceRef{
			ID: serviceRef.Status.KongID,
		},
	}

	ret, err := c.kongClient.Route.Post(kongRouteAPI)
	if err != nil {
		glog.V(4).Infof("Return from kong: Error Code=%d, %v", ret.StatusCode, ret.Body)
		return kgr, err
	}
	kgcr, err := unmarshalRoute(ret.Body)
	if err != nil {
		return kgr, err
	}
	return c.updateStatus(kgr, kgcr)
}

func (c *Controller) updateStatus(kgr *apimv1alpha1.KongRoute, kgcr *kongClient.Route) (*apimv1alpha1.KongRoute, error) {
	kgrCopy := kgr.DeepCopy()
	kgrCopy.Status = apimv1alpha1.KongRouteStatus{
		KongStatus:   "Registered",
		KongID:       kgcr.ID,
		ServiceRefID: kgcr.Service.ID,
		CreationDate: utils.UnixTimeStr(kgcr.CreationDate),
		UpdateDate:   utils.UnixTimeStr(kgcr.UpdateDate),
	}
	return c.apimClientset.ApimV1alpha1().KongRoutes(kgrCopy.Namespace).Update(kgrCopy)
}

func unmarshalRoute(resp io.ReadCloser) (*kongClient.Route, error) {
	body, err := ioutil.ReadAll(resp)
	if err != nil {
		return nil, err
	}
	var kcgr = new(kongClient.Route)
	err = json.Unmarshal([]byte(body), &kcgr)
	if err != nil {
		return nil, err
	}
	glog.V(5).Infof("Unmarshal struct => %v", kcgr)
	return kcgr, nil
}
