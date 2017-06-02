# KubeReflectorTest
A toy example about the usage of Kubernetes Reflector.

## Kubernetes Reflector ##

Reflector is a key component for Kubernetes clients, kube-scheduler and Replication Controller.
Reflector is in the middle between client and Kubernetes API-server. It provides a framework to monitor the changes of the Kubernetes cluster.

### Definition of Reflector ###
Here is the definition of Reflector.
![reflector define](https://cloud.githubusercontent.com/assets/27221807/26737893/1bc26ccc-479a-11e7-8291-f3551d5c2e6c.png)

As shown in the definition, there are two important compoents of a Reflector:
 - **ListerWatcher**
 
   It provides two functions: *List* and *Watch*. These two functions will talk with Kubernetes API-server, and get the Events.
   
 - **Store**
 
   It is usually a in-memory storage, such as HashMap, or Queue(for FIFO). Reflector will add the Events into this Store.
   
It should be noted that the **reflect.Type** is usually a Kind of Kubernetes Object, such as Pod, Node.


### Toy Example Usage of Reflector ###
Before going deeper into the implementation of Reflector, let's have a look at how Reflector can be used from a [example](https://github.com/songbinliu/KubeReflectorTest).
This example will watch the changes(including ADD, DELETE, MODIFY) of all the Pods in the Kubernetes cluster. According to the changes, it will add(or update) Pod object into the *Store* if a Pod is created(or updated), and delete a Pod object from the *Store* if the Pod is killed in the Kubernetes.  In addition, it will also print all the Pod names every 30 seconds: If a Pod is deleted, then its name won't appear; if a Pod is created, its name will appear.

```go
func main() {
	client := getKubeClient()
	if client == nil {
		fmt.Println("failed to get kubeclient")
		return
	}

	stopCh := make(chan struct{})
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)

	selector := fields.Everything()
	namespaceAll := ""
	listWatch := cache.NewListWatchFromClient(client.CoreV1Client.RESTClient(),
		"pods",
		namespaceAll,
		selector)

	r := cache.NewReflector(listWatch, &v1.Pod{}, store, 0)

	r.RunUntil(stopCh)

	for i := 1; i < 20; i++ {
		time.Sleep(30 * time.Second)
		printContent(store.ListKeys())
	}

	time.Sleep(10 * time.Second)
	printContent(store.ListKeys())
	close(stopCh)
}
 
```

From this example, we can see that Reflector uses Kube-client to list and watch all the Pods, and store the changes in the *Store*.
![reflector](https://cloud.githubusercontent.com/assets/27221807/26739964/d1db6246-47a1-11e7-8639-49699e75132e.png)


### Change For Visibility ###
For easier understanding of the Reflector, it is better to add some logs to the **Reflector.watchHandler()** function.
```go
func (r *Reflector) watchHandler(w watch.Interface, resourceVersion *string, errc chan error, stopCh <-chan struct{}) error {
	start := r.clock.Now()
	eventCount := 0
	glog.V(1).Infof("%s: begin watchHandler() *************", r.name)
	// Stopping the watcher should be idempotent and if we return from this function there's no way
	// we're coming back in with the same watch interface.
	defer w.Stop()

loop:
	for {
		select {
		case <-stopCh:
			return errorStopRequested
		case err := <-errc:
			return err
		case event, ok := <-w.ResultChan():
			if !ok {
				break loop
			}
			glog.V(1).Infof("%s: got event: %v", r.name, event.Type)
			if event.Type == watch.Error {
				return apierrs.FromObject(event.Object)
			}
```

The line ```glog.V(1).Infof("%s: got event: %v", r.name, event.Type)``` is added.
