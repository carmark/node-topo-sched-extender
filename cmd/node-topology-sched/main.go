package main

import (
	"flag"
	"net/http"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"github.com/gpucloud/node-topology-manager/pkg/controller"
	"github.com/gpucloud/node-topology-manager/pkg/routes"
	"github.com/gpucloud/node-topology-manager/pkg/scheduler"
	"github.com/gpucloud/node-topology-manager/pkg/signals"
	"github.com/julienschmidt/httprouter"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	informerFactory := informers.NewSharedInformerFactory(kubeClient, 30*time.Second)
	controller, err := controller.NewController(kubeClient, informerFactory, stopCh)
	if err != nil {
		klog.Fatalf("Failed to start due to %v", err)
	}

	go controller.Run(2, stopCh)

	topoPriority := scheduler.NewTopoSchedulerPriority("topo-scheduler", kubeClient, controller.GetSchedulerCache())

	router := httprouter.New()
	routes.AddPriority(router, topoPriority)

	klog.Infof("server starting on the port :3767")
	if err := http.ListenAndServe(":3767", router); err != nil {
		klog.Fatal(err)
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
