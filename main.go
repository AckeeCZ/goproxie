package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AckeeCZ/goproxie/internal/gcloud"
	"github.com/AckeeCZ/goproxie/internal/history"
	"github.com/AckeeCZ/goproxie/internal/kubectl"
	"github.com/AckeeCZ/goproxie/internal/store"
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
	/** Dont save to history */
	noSave *bool
}

var flags = &Flags{}

type selectFieldOption struct {
	title string
	value interface{}
}
type selectField struct {
	titleChoose  string
	titleLoading string
	valueTitle   string
	getOptions   func() []selectFieldOption
}

func promptSelection(sel selectField) interface{} {
	// Load options
	loadingStart(fmt.Sprintf("Loading %v", sel.titleLoading))
	options := sel.getOptions()
	loadingStop()
	// Serialize options to strings
	optionTitles := []string{}
	for _, option := range options {
		optionTitles = append(optionTitles, option.title)
	}
	pickedTitle := ""
	if sel.valueTitle != "" {
		// Apply selection, if set
		filtered := filterStrings(optionTitles, sel.valueTitle)
		if len(filtered) > 0 {
			pickedTitle = filtered[0]
			fmt.Printf("%v: %v\n", sel.titleChoose, pickedTitle)
		}
	} else {
		// Pick from Input otherwise
		prompt := &survey.Select{
			Message: fmt.Sprintf("Choose %v:", sel.titleChoose),
			Options: optionTitles,
		}
		survey.AskOne(prompt, &pickedTitle)
	}
	var pickedOption selectFieldOption
	// Reverse-search Option by picked title
	for _, option := range options {
		if option.title == pickedTitle {
			pickedOption = option
		}
	}
	return pickedOption.value
}

func readProjectID() string {
	return promptSelection(selectField{
		titleLoading: "GCP Projects",
		titleChoose:  "GCP Project",
		getOptions: func() (options []selectFieldOption) {
			for _, project := range gcloud.ProjectsList() {
				options = append(options, selectFieldOption{title: project, value: project})
			}
			return
		},
		valueTitle: *flags.project,
	}).(string)
}

func readCluster(projectID string) *gcloud.Cluster {
	return promptSelection(selectField{
		titleLoading: "Clusters",
		titleChoose:  "Cluster",
		getOptions: func() (options []selectFieldOption) {
			for _, cluster := range gcloud.ContainerClustersList(projectID) {
				options = append(options, selectFieldOption{title: cluster.Name, value: cluster})
			}
			return
		},
		valueTitle: *flags.cluster,
	}).(*gcloud.Cluster)
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
	return promptSelection(selectField{
		titleLoading: "K8S Namespaces",
		titleChoose:  "K8S Namespace",
		getOptions: func() (options []selectFieldOption) {
			for _, namespace := range kubectl.NamespacesList() {
				options = append(options, selectFieldOption{title: namespace, value: namespace})
			}
			return
		},
		valueTitle: *flags.namespace,
	}).(string)
}

func readPod(namespace string) *kubectl.Pod {
	return promptSelection(selectField{
		titleLoading: "Pods",
		titleChoose:  "Pod",
		getOptions: func() (options []selectFieldOption) {
			for _, pod := range kubectl.PodsList(namespace) {
				options = append(options, selectFieldOption{title: pod.Name, value: pod})
			}
			return
		},
		valueTitle: *flags.pod,
	}).(*kubectl.Pod)
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
	return promptSelection(selectField{
		titleLoading: "Remote ports",
		titleChoose:  "Remote port",
		getOptions: func() (options []selectFieldOption) {
			for _, port := range containerPorts {
				options = append(options, selectFieldOption{title: strconv.Itoa(port), value: port})
			}
			return
		},
	}).(int)
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
	flags.noSave = flag.Bool("no-save", false, "Don't save invocation to history")
	flag.Parse()
	gcloud.SetGcloudPath(*gcloudPath)
	kubectl.SetKubectlPath(*kubectlPath)
}

func main() {
	readArguments()
	store.Initialize()
	if len(os.Args) > 1 && os.Args[1] == "history" {
		history.Browse()
		return
	}
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
		if *flags.noSave == false {
			history.StorePodProxy(projectID, cluster, namespace, pod, localPort, remotePort)
		}
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
