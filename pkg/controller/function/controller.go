package function

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	clientset "github.com/dajac/kfn/pkg/client/clientset/versioned"
	informers "github.com/dajac/kfn/pkg/client/informers/externalversions/kfn/v1alpha1"
	listers "github.com/dajac/kfn/pkg/client/listers/kfn/v1alpha1"
	"github.com/golang/glog"
)

type Controller struct {
	kubeClient        kubernetes.Interface
	kfnClient         clientset.Interface
	deployementLister appslisters.DeploymentLister
	deployementSynced cache.InformerSynced
	configMapLister   corelisters.ConfigMapLister
	configMapSynched  cache.InformerSynced
	functionLister    listers.FunctionLister
	functionSynced    cache.InformerSynced

	functionDefaultConfig FunctionDefaultConfig

	workqueue workqueue.RateLimitingInterface
}

func NewController(
	kubeClient kubernetes.Interface,
	kfnClient clientset.Interface,
	deployementInformer appsinformers.DeploymentInformer,
	configMapInformer coreinformers.ConfigMapInformer,
	functionInformer informers.FunctionInformer,
	functionBaseConfig FunctionDefaultConfig) *Controller {

	controller := &Controller{
		kubeClient:            kubeClient,
		kfnClient:             kfnClient,
		deployementLister:     deployementInformer.Lister(),
		deployementSynced:     deployementInformer.Informer().HasSynced,
		configMapLister:       configMapInformer.Lister(),
		configMapSynched:      configMapInformer.Informer().HasSynced,
		functionLister:        functionInformer.Lister(),
		functionSynced:        functionInformer.Informer().HasSynced,
		functionDefaultConfig: functionBaseConfig,
		workqueue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Functions"),
	}

	deployementInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	configMapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	functionInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueFunction,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueFunction(new)
		},
		DeleteFunc: controller.enqueueFunction,
	})

	return controller
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	glog.Info("Starting Function controller")

	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deployementSynced, c.configMapSynched, c.functionSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}

		c.workqueue.Forget(obj)

		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	glog.Infof("Synching %s/%s", namespace, name)

	function, err := c.functionLister.Functions(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("Function '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	functionConfig := newFunctionConfig(&c.functionDefaultConfig, function)

	configmap, err := c.configMapLister.ConfigMaps(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			glog.Info("Create ConfigMap for %s/%s", namespace, name)
			configmap, err = c.kubeClient.CoreV1().ConfigMaps(namespace).Create(newConfigMap(function, functionConfig))
		}
	} else {
		newConfigMap := newConfigMap(function, functionConfig)

		curHash := hash(configmap)
		newHash := hash(newConfigMap)

		if curHash != newHash {
			glog.Info("Update ConfigMap for %s/%s", namespace, name)
			configmap, err = c.kubeClient.CoreV1().ConfigMaps(namespace).Update(newConfigMap)
		}
	}

	if err != nil {
		return err
	}

	deployement, err := c.deployementLister.Deployments(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			glog.Info("Create Deployement for %s/%s", namespace, name)
			deployement, err = c.kubeClient.AppsV1().Deployments(namespace).Create(newDeployement(function, configmap))
		}
	} else {
		newDeployement := newDeployement(function, configmap)

		if *newDeployement.Spec.Replicas != *deployement.Spec.Replicas || newDeployement.Spec.Template.Spec.Containers[0].Image != deployement.Spec.Template.Spec.Containers[0].Image || newDeployement.Spec.Template.Annotations["kfn.dajac.io/config-hash"] != deployement.Spec.Template.Annotations["kfn.dajac.io/config-hash"] {
			glog.Info("Update Deployement for %s/%s", namespace, name)
			deployement, err = c.kubeClient.AppsV1().Deployments(namespace).Update(newDeployement)
		}
	}

	if err != nil {
		return err
	}

	newFunction := function.DeepCopy()
	newFunction.Status.ObservedGeneration = function.Generation
	newFunction.Status.AvailableReplicas = deployement.Status.AvailableReplicas

	// Update the status only if it has changed
	if function.Status.AvailableReplicas == newFunction.Status.AvailableReplicas &&
		function.Status.ObservedGeneration == newFunction.Status.ObservedGeneration {
		return nil
	}

	_, err = c.kfnClient.Kfn().Functions(namespace).UpdateStatus(newFunction)

	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) enqueueFunction(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}

	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		if ownerRef.Kind != "Function" {
			return
		}

		function, err := c.functionLister.Functions(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			glog.V(4).Infof("ignoring orphaned object '%s' of Function '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueFunction(function)
		return
	}
}
