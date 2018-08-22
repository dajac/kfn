package controller

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

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting Foo controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deployementSynced, c.configMapSynched, c.functionSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process Foo resources
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

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
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

	glog.Infof("%+v", function)
	glog.Infof("%+v", functionConfig)

	configmap, err := c.configMapLister.ConfigMaps(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create ConfigMap
			glog.Info("Create ConfigMap")
			configmap, err = c.kubeClient.CoreV1().ConfigMaps(namespace).Create(newConfigMap(function, functionConfig))
		}
	} else {
		glog.Info("Update ConfigMap")
		newConfigMap := newConfigMap(function, functionConfig)

		curHash := hash(configmap)
		newHash := hash(newConfigMap)

		glog.Infof("Current: %s, New: %s", curHash, newHash)

		// Update ConfigMap if the content has changed
		if curHash != newHash {
			configmap, err = c.kubeClient.CoreV1().ConfigMaps(namespace).Update(newConfigMap)
		}
	}

	glog.Infof("%+v", configmap)

	if err != nil {
		return err
	}

	deployement, err := c.deployementLister.Deployments(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Deployement
			glog.Info("Create Deployement")
			deployement, err = c.kubeClient.AppsV1().Deployments(namespace).Create(newDeployement(function, configmap))
		}
	} else {
		glog.Info("Update Deployement")
		newDeployement := newDeployement(function, configmap)

		if *newDeployement.Spec.Replicas != *deployement.Spec.Replicas || newDeployement.Spec.Template.Spec.Containers[0].Image != deployement.Spec.Template.Spec.Containers[0].Image || newDeployement.Spec.Template.Annotations["kfn.dajac.io/config-hash"] != deployement.Spec.Template.Annotations["kfn.dajac.io/config-hash"] {
			deployement, err = c.kubeClient.AppsV1().Deployments(namespace).Update(newDeployement)
		}
	}

	glog.Infof("%+v", deployement)

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
	glog.Infof("Enqueue Function %s", key)
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
	glog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		if ownerRef.Kind != "Function" {
			return
		}

		function, err := c.functionLister.Functions(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			glog.V(4).Infof("ignoring orphaned object '%s' of Function '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		glog.Infof("Enqueue Function %s/%s due to %s/%s", function.Namespace, function.Name, object.GetNamespace(), object.GetName())

		c.enqueueFunction(function)
		return
	}
}
