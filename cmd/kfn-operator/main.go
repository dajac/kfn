package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/dajac/kfn/pkg/client/clientset/versioned"
	informers "github.com/dajac/kfn/pkg/client/informers/externalversions"
	"github.com/dajac/kfn/pkg/controller"
)

var (
	masterURL  string
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func main() {
	flag.Parse()

	glog.Info("Starting kfn controller")

	stopCh := make(chan struct{})
	defer close(stopCh)

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	kfnClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kfn clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	kfnInformerFactory := informers.NewSharedInformerFactory(kfnClient, time.Second*30)

	// TODO: Create controller
	controller := controller.NewController(
		kubeClient,
		kfnClient,
		kubeInformerFactory.Apps().V1().Deployments(),
		kubeInformerFactory.Core().V1().ConfigMaps(),
		kfnInformerFactory.Kfn().V1alpha1().Functions(),
	)

	go kubeInformerFactory.Start(stopCh)
	go kfnInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	glog.Info("Shutting down")
}
