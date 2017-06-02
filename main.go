package main

import (
	"flag"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func printPods(pods *v1.PodList) {
	fmt.Printf("api version:%s, kind:%s, r.version:%s\n",
		pods.APIVersion,
		pods.Kind,
		pods.ResourceVersion)

	for _, pod := range pods.Items {
		fmt.Printf("%s/%s, phase:%s, cluster:%s, host:%s\n",
			pod.Namespace,
			pod.Name,
			pod.Status.Phase,
			pod.ClusterName,
			pod.Status.HostIP)
	}
}

func testPod(client *kubernetes.Clientset) {
	pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	printPods(pods)
}

func getKubeClient() *kubernetes.Clientset {
	masterurl := flag.String("masterUrl", "", "master url")
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")

	flag.Parse()

	if *masterurl == "" && *kubeconfig == "" {
		fmt.Println("must specify masterUrl or kubeconfig.")
		return nil
	}

	var err error
	var config *restclient.Config

	if *kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		config, err = clientcmd.BuildConfigFromFlags(*masterurl, "")
	}

	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	testPod(clientset)
	return clientset
}

func printContent(arr []string) {
	fmt.Printf(" there are %d items\n", len(arr))
	for _, pod := range arr {
		fmt.Println(pod)
	}
}

func main() {
	client := getKubeClient()
	if client == nil {
		fmt.Println("failed to get kubeclient")
		return
	}

	stopCh := make(chan struct{})
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)

	//selector := fields.SelectorFromSet(nil)
	selector := fields.Everything()
	namespaceAll := ""
	listWatch := cache.NewListWatchFromClient(client.CoreV1Client.RESTClient(),
		"pods",
		namespaceAll,
		selector)

	cycle := time.Millisecond * 0
	r := cache.NewReflector(listWatch, &v1.Pod{}, store, cycle)

	r.RunUntil(stopCh)

	for i := 1; i < 20; i++ {
		time.Sleep(30 * time.Second)
		printContent(store.ListKeys())
	}

	time.Sleep(10 * time.Second)
	printContent(store.ListKeys())
	close(stopCh)
}
