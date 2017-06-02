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

### Implementation of Reflector ###
The implementation of Reflector is in [Kubernetes Go Client](https://github.com/kubernetes/client-go/blob/master/tools/cache/reflector.go) tools/cache/ package. **Reflector.RunUntil()** function will periodly call **Reflector.ListAndWatch()** function. 

As the name suggests, **Reflector.ListAndWatch()** will first list all the Objects and store them in the *Store*; then it will *watch* the changes and handle the changes with **Reflector.watchHandler()** function.Following is the skech of the **Reflector.ListAndWatch()** and **Reflector.watchHandler()** functions.

**Reflctor.ListAndWatch()**
![listwatch](https://cloud.githubusercontent.com/assets/27221807/26740901/65b646f4-47a5-11e7-9e2b-24e78d3bb65e.png)

As shown in the definition of *Reflctor.ListAndWatch()*, Reflector first call *kubeclient.List()* to get all the Objects (in the previous example, they are Pods), and add all these Objects into the *Store* by **Reflector.syncWith()**. 
Second, the Reflector will call *kubeclient.Watch()*. As this *Watch()* call has timeout, it will call *kubeclient.Watch()* repeatedly.


**Reflector.watchHandler()** 
![handlewatch](https://cloud.githubusercontent.com/assets/27221807/26740995/c2eefa00-47a5-11e7-8203-4d8e6efdb21d.png)

As shown in the definition of *Reflector.watchHandler()*, Reflector keeps a connection to the *APIserver*, and receives changes(*Events*) from this connection. It will update the content of the *Store* according to the *Event.Type*. The content  of the *Store* will be consumed by other components. 

It should be noted the *Event.Type* is not add to *Store* directly. It can be added by implementing a particular *Store*. For example, [**Delta_FIFO**](https://github.com/kubernetes/client-go/blob/master/tools/cache/delta_fifo.go) adds the *Event.Type* back according to function type, for example:
```go
// Update is just like Add, but makes an Updated Delta.
func (f *DeltaFIFO) Update(obj interface{}) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.populated = true
	//Update is the Event.Type
	return f.queueActionLocked(Updated, obj)
}

```


### Usage of Reflector ###
**Reflector** provides an efficient framework to monitor the changes of the Kubernetes cluster. Many other tools are build base on **Reflector**, by consuming the content of the *Reflector.Store*. For example,[**Controller**](https://github.com/kubernetes/client-go/blob/master/tools/cache/controller.go) has a **ProcessFunc** to consume the content of the store.

