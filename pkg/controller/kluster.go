package controller

import (
	"log"
	"time"

	klientset "github.com/vikas-gautam/kluster/pkg/client/clientset/versioned"
	kinf "github.com/vikas-gautam/kluster/pkg/client/informers/externalversions/golearning.dev/v1alpha1"
	klister "github.com/vikas-gautam/kluster/pkg/client/listers/golearning.dev/v1alpha1"
	"github.com/vikas-gautam/kluster/pkg/do"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	//k8s native client
	k8sclient kubernetes.Interface

	//clientset for custom resource kluster
	klient klientset.Interface

	//kluster has synced
	klusterSynced cache.InformerSynced

	//lister
	klister klister.KlusterLister

	//queue
	wq workqueue.RateLimitingInterface
}

func NewController(k8sclient kubernetes.Interface, klient klientset.Interface, klusterInformer kinf.KlusterInformer) *Controller {
	//to initialize controller we need clientset and informer of custom type

	c := &Controller{
		k8sclient:     k8sclient,
		klient:        klient,
		klusterSynced: klusterInformer.Informer().HasSynced,
		klister:       klusterInformer.Lister(),
		wq:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kluster"),
	}

	klusterInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDel,
		},
	)
	return c
}

func (c *Controller) Run(ch chan struct{}) error {
	//before running the controller, ensure the local cache in informers has been initialized atleast once
	if ok := cache.WaitForCacheSync(ch, c.klusterSynced); !ok {
		log.Println("cache was not synced")
	}

	//run goroutine that is going to consume object from this workqueue continously
	go wait.Until(c.worker, time.Second, ch)

	<-ch
	return nil
}

func (c *Controller) worker() {
	for c.processNextItem() {

	}
}

func (c *Controller) processNextItem() bool {
	//get the resource from the cache
	item, shutdown := c.wq.Get()
	if shutdown {
		//logs as well
		return false
	}
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Printf("error %s calling namespace key func on cache for item\n", err.Error())
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("splitting key into naemspace and name, error %s", err.Error())
		return false
	}

	//we will use lister instead of calling api server
	kluster, err := c.klister.Klusters(ns).Get(name)
	if err != nil {
		log.Printf("getting the kluster resource from lister  %s", err.Error())
		return false
	}
	log.Printf("kluster spec that we have is %+v\n", kluster.Spec)

	clusterID, err := do.Create(c.k8sclient, kluster.Spec)
	if err != nil {
		log.Printf("error in creating cluster %s", err.Error())
	}
	log.Printf("cluster ID that we have created %s\n", clusterID) 

	return true
}

func (c *Controller) handleAdd(obj interface{}) {
	log.Println("handleAdd was called")
	c.wq.Add(obj)

}

func (c *Controller) handleDel(obj interface{}) {
	log.Println("handleDel was called")
	c.wq.Add(obj)

}
