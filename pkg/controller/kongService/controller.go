package kongService

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"time"

	apimv1alpha1 "github.com/cdiscount/kong-operator/pkg/apis/apim/v1alpha1"
	clientset "github.com/cdiscount/kong-operator/pkg/client/clientset/versioned"
	apimscheme "github.com/cdiscount/kong-operator/pkg/client/clientset/versioned/scheme"
	informers "github.com/cdiscount/kong-operator/pkg/client/informers/externalversions"
	listers "github.com/cdiscount/kong-operator/pkg/client/listers/apim/v1alpha1"
	utils "github.com/cdiscount/kong-operator/pkg/utils"
	kongClient "github.com/etiennecoutaud/kong-client-go/kong"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const (
	maxRetries = 15
	// controllerAgentName is the name in event sources
	controllerAgentName = "kongService-controller"

	// SuccessSynced is used as part of the Event 'reason' when a Database is synced
	SuccessSynced = "Synced"
)

// Controller is the controller implementation for kanary resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	apimClientset clientset.Interface

	kongServiceLister listers.KongServiceLister

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
		kongServiceLister:       kongServiceInformer.Lister(),
		kongServiceListerSynced: kongServiceInformer.Informer().HasSynced,
		workqueue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kongService"),
		recorder:                recorder,
		kongClient:              kongClient,
	}

	kongServiceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addKongServiceHandler,
		UpdateFunc: controller.updateKongServiceHandler,
		DeleteFunc: controller.deleteKongServiceHandler,
	})

	glog.Info("Setting up event handlers")
	return controller
}

func (c *Controller) addKongServiceHandler(obj interface{}) {
	kongService := obj.(*apimv1alpha1.KongService)
	glog.Info("Add KongService ", kongService.Name)
	c.enqueue(kongService)
}

func (c *Controller) updateKongServiceHandler(old interface{}, cur interface{}) {
	oldKongService := old.(*apimv1alpha1.KongService)
	kongService := cur.(*apimv1alpha1.KongService)
	if !(reflect.DeepEqual(oldKongService.Spec, kongService.Spec)) {
		glog.Info("Update KongService:", kongService.Name)
		c.enqueue(kongService)
	}
}

func (c *Controller) deleteKongServiceHandler(obj interface{}) {
	kongService := obj.(*apimv1alpha1.KongService)
	//Delete only if object has been created
	if !reflect.DeepEqual(kongService.Status, apimv1alpha1.KongServiceStatus{}) {
		err := c.deleteService(kongService)
		if err != nil {
			glog.Fatalf("Error when delete KongService: %s => %s", kongService.Name, err)
		}
		glog.Infof("KongService: %s deleted", kongService.Name)
	}
}

// enqueue takes a resource and converts it into a namespace/name
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
	glog.Info("Starting KongService controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.kongServiceListerSynced); !ok {
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
		glog.V(2).Infof("Error syncing kongService %v: %v", key, err)
		c.workqueue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	glog.V(2).Infof("Dropping kongService %q out of the queue: %v", key, err)
	c.workqueue.Forget(key)
}

// Sync loop for kongService resources
func (c *Controller) reconcile(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the KongService resource with this namespace/name
	kgs, err := c.kongServiceLister.KongServices(namespace).Get(name)
	if err != nil {
		// processing.
		if errors.IsNotFound(err) {
			glog.V(2).Info(fmt.Sprintf("KongService %s no longer exists", key))
			return nil
		}
		return err
	}

	// Create new KongService
	// If no status then new object
	// Else need to update
	if reflect.DeepEqual(kgs.Status, apimv1alpha1.KongServiceStatus{}) {
		kgs, err = c.createService(kgs)
		if err != nil {
			c.recorder.Event(kgs, corev1.EventTypeNormal, "Error", err.Error())
			return err
		}
		c.recorder.Event(kgs, corev1.EventTypeNormal, "Info", "Kong Service created")
	} else {
		kgs, err = c.updateService(kgs)
		if err != nil {
			c.recorder.Event(kgs, corev1.EventTypeNormal, "Error", err.Error())
			return err
		}
		c.recorder.Event(kgs, corev1.EventTypeNormal, "Info", "Kong Service updated")
	}

	c.recorder.Event(kgs, corev1.EventTypeNormal, "Sync", "Correctly sync")
	return nil
}

func (c *Controller) updateService(kgs *apimv1alpha1.KongService) (*apimv1alpha1.KongService, error) {
	kongServiceAPI := &kongClient.Service{
		Name:           kgs.Name,
		Protocol:       kgs.Spec.Protocol,
		Host:           kgs.Spec.Host,
		Port:           kgs.Spec.Port,
		Path:           kgs.Spec.Path,
		Retries:        kgs.Spec.Retries,
		ConnectTimeout: kgs.Spec.ConnectTimeout,
		WriteTimeout:   kgs.Spec.WriteTimeout,
		ReadTimeout:    kgs.Spec.ReadTimeout,
	}
	ret, err := c.kongClient.Service.Patch(kgs.Status.KongID, kongServiceAPI)
	if err != nil {
		glog.V(4).Infof("Return from kong: Error Code=%d, %v", ret.StatusCode, ret.Body)
		return kgs, err
	}
	kgcs, err := unmarshalService(ret.Body)
	if err != nil {
		return kgs, err
	}
	return c.updateStatus(kgs, kgcs)
}

func (c *Controller) createService(kgs *apimv1alpha1.KongService) (*apimv1alpha1.KongService, error) {
	kongServiceAPI := &kongClient.Service{
		Name:           kgs.Name,
		Protocol:       kgs.Spec.Protocol,
		Host:           kgs.Spec.Host,
		Port:           kgs.Spec.Port,
		Path:           kgs.Spec.Path,
		Retries:        kgs.Spec.Retries,
		ConnectTimeout: kgs.Spec.ConnectTimeout,
		WriteTimeout:   kgs.Spec.WriteTimeout,
		ReadTimeout:    kgs.Spec.ReadTimeout,
	}

	ret, err := c.kongClient.Service.Post(kongServiceAPI)
	if err != nil {
		glog.V(4).Infof("Return from kong: Error Code=%d, %v", ret.StatusCode, ret.Body)
		return kgs, err
	}
	kgcs, err := unmarshalService(ret.Body)
	if err != nil {
		return kgs, err
	}
	return c.updateStatus(kgs, kgcs)
}

func (c *Controller) updateStatus(kgs *apimv1alpha1.KongService, kgcs *kongClient.Service) (*apimv1alpha1.KongService, error) {
	kgsCopy := kgs.DeepCopy()
	kgsCopy.Status = apimv1alpha1.KongServiceStatus{
		KongStatus:   "Registered",
		KongID:       kgcs.ID,
		URL:          kgs.Spec.Host + ":" + strconv.Itoa(kgs.Spec.Port) + kgs.Spec.Path,
		CreationDate: utils.UnixTimeStr(kgcs.CreationDate),
		UpdateDate:   utils.UnixTimeStr(kgcs.UpdateDate),
	}
	return c.apimClientset.ApimV1alpha1().KongServices(kgsCopy.Namespace).Update(kgsCopy)
}

func unmarshalService(resp io.ReadCloser) (*kongClient.Service, error) {
	body, err := ioutil.ReadAll(resp)
	if err != nil {
		return nil, err
	}
	var kcgs = new(kongClient.Service)
	err = json.Unmarshal([]byte(body), &kcgs)
	if err != nil {
		return nil, err
	}
	glog.V(5).Infof("Unmarshal struct => %v", kcgs)
	return kcgs, nil
}

func (c *Controller) deleteService(kgs *apimv1alpha1.KongService) error {
	_, err := c.kongClient.Service.Delete(kgs.Status.KongID)
	return err
}
