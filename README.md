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
./controller-gen paths=github.com/vikas-gautam/kluster/apis/golearning.dev/v1alpha1  crd:crdVersions=v1 output:crd:artifacts:config=manifests 
```
create cr in manifests folder

#subresources and additional printer columns 
once cluster is created on DO it returns ClusterID
now to proceed further we need to store ClusterID somewhere

Hence we can hava status field appended to the yaml (by controller) of custom resource which 
will have all the required information.

apiVersion:
kind:
metadata:
spec:
    ---
    ---
status:
    clusterid
    kubeconfig
    progress

status is written by controller only.


#subresources:
pod - resource
logs as subresources (native subresources)

------------------------------------------------------------------------------------------------

#event recorder for kluster and routines to handle objects from queue: P5

progress status field is static as of now i.e. creating
in kluster.go - run a DO API query to check cluster status and then change that static - creating state to runnng
waitForCluster to be up and running, once it is running update status to be running

record events - it prints the events under events field when u describe resource.
all the controllers or components those do something for resource they update their events under
event field of that particular resource.








