package main

import (
	"flag"
	"log"
	"path/filepath"
	"time"

	klient "github.com/vikas-gautam/kluster/pkg/client/clientset/versioned"
	kInfFac "github.com/vikas-gautam/kluster/pkg/client/informers/externalversions"
	"github.com/vikas-gautam/kluster/pkg/controller"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {

	// k := v1alpha1.Kluster{}
	// fmt.Println(k)

	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Building config from flags failed, %s, trying to build inclusterconfig", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("error %s building inclusterconfig", err.Error())
		}
	}
	klientset, err := klient.NewForConfig(config)
	if err != nil {
		log.Printf("getting klient set %s\n", err.Error())
	}
	// fmt.Println(klientset)

	// klist, err := klientset.GolearningV1alpha1().Klusters("").List(context.Background(), metav1.ListOptions{})
	// if err != nil {
	// 	log.Printf("listing klusters %s\n", err.Error())
	// }

	// fmt.Println(klist.Items)

	//above code not required bcz we will call controller

	infoFactory := kInfFac.NewSharedInformerFactory(klientset, 20*time.Minute)
	ch := make(chan struct{})
	c := controller.NewController(klientset, infoFactory.Golearning().V1alpha1().Klusters())
	infoFactory.Start(ch)
	if err := c.Run(ch); err != nil{
		log.Printf("error running controller %s\n", err.Error())
	}

}
