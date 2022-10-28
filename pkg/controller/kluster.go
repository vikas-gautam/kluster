package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kanisterio/kanister/pkg/poll"
	"github.com/vikas-gautam/kluster/pkg/apis/golearning.dev/v1alpha1"
	klientset "github.com/vikas-gautam/kluster/pkg/client/clientset/versioned"
	kinf "github.com/vikas-gautam/kluster/pkg/client/informers/externalversions/golearning.dev/v1alpha1"
	klister "github.com/vikas-gautam/kluster/pkg/client/listers/golearning.dev/v1alpha1"
	"github.com/vikas-gautam/kluster/pkg/do"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	//call updateStatus method
	fmt.Println("calling updateStatus function")
	err = c.updateStatus(clusterID, "creating", kluster)
	if err != nil {
		log.Printf("error %s, updating status of the kluster %s\n", err.Error(), kluster.Name)
	}

	fmt.Println("calling the DO API to check cluster is running")
	// query the DO cluster API, make sure cluster is running
	err = c.waitForCluster(kluster.Spec, clusterID)
	if err != nil {
		log.Printf("error %s, waiting for cluster to be running", err.Error())
	}

	fmt.Println("updating the cluster status with progress running")
	//updating the kluster status
	c.updateStatus(clusterID, "running", kluster)
	if err != nil {
		log.Printf("error %s, updating cluster status after waiting for cluster status", err.Error())
	}

	return true
}

func (c *Controller) waitForCluster(spec v1alpha1.KlusterSpec, clusterID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel() 

	return poll.Wait(ctx, func(ctx context.Context) (bool, error) {
		state, err := do.ClusterState(c.k8sclient, spec, clusterID)
		if err != nil {
			return false, err
		}
		if state == "running" {
			return true, nil
		}
		return false, nil
	})
}

// updateStatus in CR i.e. kluster
func (c *Controller) updateStatus(id string, progress string, kluster *v1alpha1.Kluster) error {
	//get the latest version of resource kluster which exists after controller has added status field's values
	updatedKluster, err := c.klient.GolearningV1alpha1().Klusters(kluster.Namespace).Get(context.Background(), kluster.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	updatedKluster.Status.KlusterID = id
	updatedKluster.Status.Progress = progress
	_, err = c.klient.GolearningV1alpha1().Klusters(kluster.Namespace).UpdateStatus(context.Background(), updatedKluster, metav1.UpdateOptions{})
	return err
}

func (c *Controller) handleAdd(obj interface{}) {
	log.Println("handleAdd was called")
	c.wq.Add(obj)

}

func (c *Controller) handleDel(obj interface{}) {
	log.Println("handleDel was called")
	c.wq.Add(obj)

}
