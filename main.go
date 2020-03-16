package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/AckeeCZ/goproxie/internal/gcloud"
	"github.com/AckeeCZ/goproxie/internal/kubectl"
	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
)

func initializationCheck() {
	// TODO
}

func readProxyType() string {
	proxyType := ""
	proxyTypes := []string{"VM", "POD"}
	if *flags.proxyType != "" {
		filtered := filterStrings(proxyTypes, *flags.proxyType)
		if len(filtered) > 0 {
			proxyType = filtered[0]
			fmt.Printf("Choose proxy type: %v\n", proxyType)
		}
	} else {
		prompt := &survey.Select{
			Message: "Choose proxy type:",
			// TODO Refactor types to Enums
			Options: proxyTypes,
		}
		survey.AskOne(prompt, &proxyType)
	}
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

type Flags struct {
	project   *string
	proxyType *string
	cluster   *string
	namespace *string
	pod       *string
	localPort *string
}

var flags = &Flags{}

func readProjectID() string {
	loadingStart("Loading GCP Projects")
	projects := gcloud.ProjectsList()
	loadingStop()
	projectID := ""
	if *flags.project != "" {
		filtered := filterStrings(projects, *flags.project)
		if len(filtered) > 0 {
			projectID = filtered[0]
			fmt.Printf("Choose project: %v\n", projectID)
		}
	} else {
		prompt := &survey.Select{
			Message: "Choose project:",
			Options: projects,
		}
		survey.AskOne(prompt, &projectID)
	}

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
	if *flags.cluster != "" {
		filtered := filterStrings(clusterNames, *flags.cluster)
		if len(filtered) > 0 {
			clusterName = filtered[0]
			fmt.Printf("Choose cluster: %v\n", clusterName)
		}
	} else {
		prompt := &survey.Select{
			Message: "Choose cluster:",
			Options: clusterNames,
		}
		survey.AskOne(prompt, &clusterName)
	}
	var clusterByName *gcloud.Cluster
	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			clusterByName = cluster
		}
	}
	return clusterByName
}

func filterStrings(options []string, filter string) []string {
	if len(filter) == 0 {
		return options
	}
	results := []string{}

	// Inspired by Survey's filtering
	// https://github.com/AlecAivazis/survey/blob/59f4d6f95795f2e6b20526769ca4662ced786ccc/survey.go#L50
	filter = strings.ToLower(filter)
	for _, option := range options {
		if strings.Contains(strings.ToLower(option), filter) {
			results = append(results, option)
		}
	}
	return results
}

func readNamespace() string {
	loadingStart("Loading K8S Namespaces")
	namespaces := kubectl.NamespacesList()
	loadingStop()
	namespace := ""
	if *flags.namespace != "" {
		filtered := filterStrings(namespaces, *flags.namespace)
		if len(filtered) > 0 {
			namespace = filtered[0]
			fmt.Printf("Choose namespace: %v\n", namespace)
		}
	} else {
		prompt := &survey.Select{
			Message: "Choose namespace:",
			Options: namespaces,
		}
		survey.AskOne(prompt, &namespace)
	}
	return namespace
}

func readPod(namespace string) *kubectl.Pod {
	loadingStart("Loading Pods")
	pods := kubectl.PodsList(namespace)
	loadingStop()
	var podName string
	podOptions := make([]string, 0, len(pods))
	for _, pod := range pods {
		podOptions = append(podOptions, pod.Name)
	}
	if *flags.pod != "" {
		filtered := filterStrings(podOptions, *flags.pod)
		if len(filtered) > 0 {
			podName = filtered[0]
			fmt.Printf("Choose pod: %v\n", podName)
		}
	} else {
		prompt := &survey.Select{
			Message: "Choose pod:",
			Options: podOptions,
		}
		survey.AskOne(prompt, &podName)
	}
	var pickedPod *kubectl.Pod
	for _, pod := range pods {
		if pod.Name == podName {
			pickedPod = pod
		}
	}
	return pickedPod
}

func readLocalPort(defaultPort int) int {
	port := "3000"
	if *flags.localPort != "" {
		port = *flags.localPort
		fmt.Printf("Choose local port: %v\n", port)
	} else {
		prompt := &survey.Input{
			Message: "Choose local port:",
			Default: strconv.Itoa(defaultPort),
		}
		survey.AskOne(prompt, &port)
	}
	n, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}
	return n
}

func readRemotePort(containerPorts []int) int {
	port := "3000"
	remotePortOptions := make([]string, 0, len(containerPorts))
	for _, port := range containerPorts {
		remotePortOptions = append(remotePortOptions, strconv.Itoa(port))
	}
	prompt := &survey.Select{
		Message: "Choose remote port:",
		Options: remotePortOptions,
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
	flags.project = flag.String("project", "", "Auto GCP Project pick")
	flags.proxyType = flag.String("proxy_type", "", "Auto Proxy type pick")
	flags.cluster = flag.String("cluster", "", "Auto Cluster pick")
	flags.namespace = flag.String("namespace", "", "Auto Namespace pick")
	flags.pod = flag.String("pod", "", "Auto Pod pick")
	flags.localPort = flag.String("local_port", "", "Auto Local port pick")
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
		namespace := readNamespace()
		pod := readPod(namespace)
		remotePort := readRemotePort(pod.ContainerPorts)
		localPort := readLocalPort(remotePort)
		kubectl.PortForward(pod.Name, localPort, remotePort, namespace)
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
