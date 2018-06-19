package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"goji.io"
	"goji.io/pat"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/cdiscount/kong-operator/pkg/signal"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	clientset "github.com/cdiscount/kong-operator/pkg/client/clientset/versioned"
	apimInformers "github.com/cdiscount/kong-operator/pkg/client/informers/externalversions"
	kongServiceController "github.com/cdiscount/kong-operator/pkg/controller/kongService"
	route "github.com/cdiscount/kong-operator/pkg/route"
	kongClient "github.com/etiennecoutaud/kong-client-go/kong"
	kubeInformers "k8s.io/client-go/informers"
)

var (
	kuberconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	master      = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	version     = "No version"
	timestamp   = "0.0"
)

func debugHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s", r.Method, r.URL)
			next.ServeHTTP(w, r)
		})
}

func main() {
	flag.Parse()
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(*master, *kuberconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	apimClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building apim clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeInformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	apimInformerFactory := apimInformers.NewSharedInformerFactory(apimClient, time.Second*30)

	kongURL := os.Getenv("KONG_URL")
	if kongURL == "" {
		glog.Fatal("KONG_URL env var should be set")
		os.Exit(1)
	}
	glog.Infof("KONG_URL: %s", kongURL)
	kc, err := kongClient.NewClient(nil, kongURL)
	if err != nil {
		glog.Fatalf("Fail to create kong client : %v", err)
	}

	//kongRouteOperator := kongRouteController.NewController(kubeClient, apimClient, kubeInformerFactory, apimInformerFactory)
	kongServiceOperator := kongServiceController.NewController(kubeClient, apimClient, kubeInformerFactory, apimInformerFactory, kc)

	//go kubeInformerFactory.Start(stopCh)
	go apimInformerFactory.Start(stopCh)

	mux := goji.NewMux()
	mux.Use(debugHandler)
	mux.HandleFunc(pat.Get("/healthz"), route.Healthz)
	mux.Handle(pat.Get("/metrics"), promhttp.Handler())

	go http.ListenAndServe(":8080", mux)

	// if err = kongRouteOperator.Run(2, stopCh); err != nil {
	// 	glog.Fatalf("Error running controller: %s", err.Error())
	// }
	if err = kongServiceOperator.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
	glog.Flush()
}
