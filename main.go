package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"ackee.cz/goproxie/internal/gcloud"
	"ackee.cz/goproxie/internal/kubectl"
	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
)

func initializationCheck() {
	// TODO
}

func readProxyType() string {
	proxyType := ""
	prompt := &survey.Select{
		Message: "Choose proxy type:",
		// TODO Refactor types to Enums
		Options: []string{"VM", "POD"},
	}
	survey.AskOne(prompt, &proxyType)
	return proxyType
}

// 💡 Spinner!
var loading = spinner.New(spinner.CharSets[21], 100*time.Millisecond)

func loadingStart(suffix string) {
	loading.Start()
	loading.Suffix = fmt.Sprintf(" %v", suffix)
}
func loadingStop() {
	loading.Stop()
	loading.Suffix = ""
}

func readProjectID() string {
	loadingStart("Loading GCP Projects")
	projects := gcloud.ProjectsList()
	loadingStop()
	projectID := ""
	prompt := &survey.Select{
		Message: "Choose project:",
		Options: projects,
	}
	survey.AskOne(prompt, &projectID)
	return projectID
}

func readCluster(projectID string) *gcloud.Cluster {
	loadingStart("Loading Clusters")
	clusters := gcloud.ContainerClustersList(projectID)
	loadingStop()
	clusterName := ""
	clusterNames := make([]string, 0, len(clusters))
	for _, cluster := range clusters {
		clusterNames = append(clusterNames, cluster.Name)
	}
	prompt := &survey.Select{
		Message: "Choose cluster:",
		Options: clusterNames,
	}
	survey.AskOne(prompt, &clusterName)
	var clusterByName *gcloud.Cluster
	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			clusterByName = cluster
		}
	}
	return clusterByName
}

// Deprecated: Some reason
func readNamespace() string {
	loadingStart("Loading K8S Namespaces")
	namespaces := kubectl.NamespacesList()
	loadingStop()
	namespace := ""
	prompt := &survey.Select{
		Message: "Choose namespace:",
		Options: namespaces,
	}
	survey.AskOne(prompt, &namespace)
	return namespace
}

func readPod() *kubectl.Pod {
	loadingStart("Loading Pods")
	pods := kubectl.PodsList()
	loadingStop()
	var podName string
	podOptions := make([]string, 0, len(pods))
	for _, pod := range pods {
		podOptions = append(podOptions, pod.Name)
	}
	prompt := &survey.Select{
		Message: "Choose pod:",
		Options: podOptions,
	}
	survey.AskOne(prompt, &podName)
	var pickedPod *kubectl.Pod
	for _, pod := range pods {
		if pod.Name == podName {
			pickedPod = pod
		}
	}
	return pickedPod
}

func readLocalPort() int {
	port := "3000"
	// TODO Preference
	prompt := &survey.Input{
		Message: "Choose local port:",
	}
	survey.AskOne(prompt, &port)
	n, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}
	return n
}

func readRemotePort() int {
	// TODO Base on remote resource port
	port := "3000"
	prompt := &survey.Input{
		Message: "Choose local port:",
	}
	survey.AskOne(prompt, &port)
	n, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}
	return n
}

func readArguments() {
	gcloudPath := flag.String("gcloud_path", "gcloud", "gcloud binary path")
	kubectlPath := flag.String("kubectl_path", "kubectl", "kubectl binary path")
	flag.Parse()
	gcloud.SetGcloudPath(*gcloudPath)
	kubectl.SetKubectlPath(*kubectlPath)
}

func main() {
	readArguments()
	projectID := readProjectID()
	proxyType := readProxyType()
	cluster := readCluster(projectID)
	if proxyType == "POD" {
		loadingStart("Loading Cluster credentials")
		gcloud.GetClusterCredentials(projectID, cluster)
		loadingStop()
		pod := readPod()
		// Remove completely? If we can read the port from Pod, makes no sense for the user to edit this
		// remotePort := readRemotePort()
		// TODO Base localport choice on remote port. Remote port is usually common port the the app type
		localPort := readLocalPort()
		kubectl.PortForward(pod.Name, localPort, pod.ContainerPort, pod.Namespace)
	}

	// fmt.Println(project_id)
	// fmt.Println(proxy_type)
	// fmt.Println(cluster)
	// fmt.Println(namespace)

	// 	Pod and VM should be fairly easy.
	//	and user would must have it installed. Goproxie would then call the installed binary.
	// For PROXYTYPE=Pod
	//	TODO: Fetch Clusters for the selected GCPPROJECT
	//	TODO: Prompt user to select the Cluster {=CLUSTER}
	//	TODO: Fetch Pods for the selected GCPPROJECT and CLUSTER
	//	TODO: Prompt user to select the Pod {=POD}
	//	TODO: Prompt user for local port number
	// 	TODO (LOW): Prefill the port with the most appropriate and free port on local machine
	//		Current Node.js/proxie logic is: If it's a VM and contains `mongo` in the name,
	//		use MongoDB's native port 27017 as the starting point for our guess,
	//		if it's CloudSQL proxy type then 3306,
	//		if Pod and name contains `mongo` in the name then 27017,
	//		3000 otherwise. Then look for first free port by increasing the number - if 3306 is full, try 3307 etc.
	//		Original logic can be foud here: https://github.com/AckeeCZ/be-scripts/blob/master/src/lib/proxie.ts#L167
	//	TODO: Prompt user for remote port number
	//	TODO (LOW): Prefill by the logic above. Remote service almost always uses the default port.
	//	TODO: Create a kubectl port-forward for the given GCPPROJECT, CLUSTER, POD and ports.
	// For PROXYTYPE=VM
	//	TODO: Complete this guide.
}
