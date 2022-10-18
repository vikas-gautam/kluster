# kluster
create code to create new runtime object (we can also call it CRD as a code) and register it using schemes
clientset would not be available by default as we are registering new runtime object (that implements deepcopy, set and get group version)
informers: used to let operator know that particluar resources created or deleted (other events etc)
lister: it will get resources from informer cache

we need global and local tags too in our code

code generator will create following in pkg dir
bash /home/vikash/go/pkg/mod/k8s.io/code-generator@v0.20.4/generate-groups.sh all github.com/vikas-gautam/kluster/pkg/client  github.com/vikas-gautam/kluster/pkg/apis "golearning.dev:v1alpha1" --go-header-file /home/vikash/go/pkg/mod/k8s.io/code-generator@v0.20.4/hack/boilerplate.go.txt

#under pkg dir
deep copy object

#under client dir
clientset
informers
lister

generate crd with controller-gen

```
./controller-gen paths=./apis/golearning.dev/v1alpha1  crd:crdVersions=v1 output:crd:artifacts:config=manifests 
```
create cr in manifests folder
