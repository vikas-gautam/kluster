package do

//we have passed secretName as secretToken in kluster yaml file so we need to get the value of token
//stored in secretToken dosecret in default ns

//TODO
// 1. create do client (required token)
// 2. get token from k8s secret
// 3. call clustercreation api

import (
	"context"
	"fmt"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/vikas-gautam/kluster/pkg/apis/golearning.dev/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Create(c kubernetes.Interface, spec v1alpha1.KlusterSpec) (string, error) {
	//tokenSecret has value in ns/secretname format
	secretNameSpace := strings.Split(spec.TokenSecret, "/")[0]
	k8sSecretName := strings.Split(spec.TokenSecret, "/")[1]

	//get the token value from k8sSecret dosecret
	tokenValue, err := getToken(c, secretNameSpace, k8sSecretName)
	if err != nil {
		fmt.Printf("Unable to get token from k8sSecret %s", err.Error())
	}

	//do client with tokenValue
	client := godo.NewFromToken(tokenValue)
	fmt.Println(client)

	//call do clustercreation api
	request := &godo.KubernetesClusterCreateRequest{
		Name:        spec.Name,
		RegionSlug:  spec.Region,
		VersionSlug: spec.Version,
		NodePools: []*godo.KubernetesNodePoolCreateRequest{
			&godo.KubernetesNodePoolCreateRequest{
				Name:  spec.NodePools[0].Name,
				Size:  spec.NodePools[0].Size,
				Count: spec.NodePools[0].Count,
			},
		},
	}

	clusterCreated, _, err := client.Kubernetes.Create(context.Background(), request)
	if err != nil {
		return "", err
	}

	return clusterCreated.ID, nil
}

func ClusterState(c kubernetes.Interface, spec v1alpha1.KlusterSpec, id string) (string, error) {
	//tokenSecret has value in ns/secretname format
	secretNameSpace := strings.Split(spec.TokenSecret, "/")[0]
	k8sSecretName := strings.Split(spec.TokenSecret, "/")[1]

	//get the token value from k8sSecret dosecret
	tokenValue, err := getToken(c, secretNameSpace, k8sSecretName)
	if err != nil {
		fmt.Printf("Unable to get token from k8sSecret %s", err.Error())
	}

	//do client with tokenValue
	client := godo.NewFromToken(tokenValue)

	cluster, _, err := client.Kubernetes.Get(context.Background(), id)
	return string(cluster.Status.State), err
}

func getToken(k8sclient kubernetes.Interface, secretNameSpace, k8sSecretName string) (string, error) {
	s, err := k8sclient.CoreV1().Secrets(secretNameSpace).Get(context.Background(), k8sSecretName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	fmt.Println(s)
	return string(s.Data["token"]), nil
}
