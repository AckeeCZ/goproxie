package main

import (
	"flag"
	"os"
	"testing"

	"github.com/AckeeCZ/goproxie/internal/gcloud"
	"github.com/AckeeCZ/goproxie/internal/kubectl"
)

func mockGcloudProjectList(mockedProjects []string) func() {
	originalFn := gcloudProjectsList
	gcloudProjectsList = func() []string {
		return mockedProjects
	}
	return func() {
		gcloudProjectsList = originalFn
	}
}

func mockKubectlPodsList(mockedPods []*kubectl.Pod) func() {
	originalFn := kubectlPodsList
	kubectlPodsList = func(_ string) []*kubectl.Pod {
		return mockedPods
	}
	return func() {
		kubectlPodsList = originalFn
	}
}

func mockGcloudContainerClustersList(mockedClusters []*gcloud.Cluster) func() {
	originalFn := gcloudContainerClustersList
	gcloudContainerClustersList = func(_ string) []*gcloud.Cluster {
		return mockedClusters
	}
	return func() {
		gcloudContainerClustersList = originalFn
	}
}

func mockProxyType(proxyType string) func() {
	originalFn := readProxyType
	readProxyType = func() string {
		return proxyType
	}
	return func() {
		readProxyType = originalFn
	}
}

func mockKubcetlNamespacesList(namespaces []string) func() {
	originalFn := kubectlNamespacesList
	kubectlNamespacesList = func() []string {
		return namespaces
	}
	return func() {
		kubectlNamespacesList = originalFn
	}
}

func mockGcloudGetClusterCredentials() func() {
	originalFn := gcloudGetClusterCredentials
	gcloudGetClusterCredentials = func(_ string, _ *gcloud.Cluster) {}
	return func() {
		gcloudGetClusterCredentials = originalFn
	}
}

type PortforwardArgs struct {
	podName    string
	localPort  int
	remotePort int
	namespace  string
}

func mockKubectlPortForward() func() PortforwardArgs {
	originalFn := kubectlPortForward
	callArgs := PortforwardArgs{}
	kubectlPortForward = func(podName string, localPort int, remotePort int, namespace string) {
		callArgs.podName = podName
		callArgs.localPort = localPort
		callArgs.remotePort = remotePort
		callArgs.namespace = namespace
	}
	return func() PortforwardArgs {
		kubectlPortForward = originalFn
		return callArgs
	}
}

func mockAll(
	projects []string,
	pods []*kubectl.Pod,
	clusters []*gcloud.Cluster,
	proxyType string,
	namespace []string,
) func() {
	unmockProjects := mockGcloudProjectList(projects)
	unmockPods := mockKubectlPodsList(pods)
	unmockClusters := mockGcloudContainerClustersList(clusters)
	unmockProxyType := mockProxyType("POD")
	unmockGetCredentials := mockGcloudGetClusterCredentials()
	unmockNamespaces := mockKubcetlNamespacesList(namespace)
	return func() {
		unmockProjects()
		unmockPods()
		unmockClusters()
		unmockProxyType()
		unmockGetCredentials()
		unmockNamespaces()
	}
}

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestHappyPath(t *testing.T) {
	resetFlags()
	unmockAll := mockAll(
		[]string{"project-1"},
		[]*kubectl.Pod{
			&kubectl.Pod{Name: "pod-1", ContainerPorts: []int{1}, Containers: []string{"container-1"}},
		},
		[]*gcloud.Cluster{
			&gcloud.Cluster{Name: "cluster-1", Location: "location-1"},
		},
		"POD",
		[]string{"namespace-1"},
	)
	defer unmockAll()
	unmockPortForward := mockKubectlPortForward()
	// Uses single choice auto-selection
	os.Args = []string{"goproxie", "-local_port=1234"}
	main()
	calledWith := unmockPortForward()
	if calledWith.localPort != 1234 {
		t.Errorf("Expected port-forward to be called with localPort=%v, but was called with %v", 1234, calledWith.localPort)
	}
	if calledWith.remotePort != 1 {
		t.Errorf("Expected port-forward to be called with remotePort=%v, but was called with %v", 1, calledWith.remotePort)
	}
	if calledWith.podName != "pod-1" {
		t.Errorf("Expected port-forward to be called with podName=%v, but was called with %v", 1, calledWith.podName)
	}
	if calledWith.namespace != "namespace-1" {
		t.Errorf("Expected port-forward to be called with namespace=%v, but was called with %v", 1, calledWith.namespace)
	}
}

func ExampleNoProjects() {
	resetFlags()
	unmockAll := mockAll(
		[]string{},
		[]*kubectl.Pod{
			&kubectl.Pod{Name: "pod-1", ContainerPorts: []int{1}, Containers: []string{"container-1"}},
		},
		[]*gcloud.Cluster{
			&gcloud.Cluster{Name: "cluster-1", Location: "location-1"},
		},
		"POD",
		[]string{"namespace-1"},
	)
	defer unmockAll()
	os.Args = []string{"goproxie", "-local_port=1234"}
	main()
	// Output: Could not find any GCP Projects
}

func ExampleNoClusters() {
	resetFlags()
	unmockAll := mockAll(
		[]string{"project-1"},
		[]*kubectl.Pod{
			&kubectl.Pod{Name: "pod-1", ContainerPorts: []int{1}, Containers: []string{"container-1"}},
		},
		[]*gcloud.Cluster{},
		"POD",
		[]string{"namespace-1"},
	)
	defer unmockAll()
	os.Args = []string{"goproxie", "-local_port=1234"}
	main()
	// Output:
	// GCP Project: project-1
	// Could not find any GCP Clusters
}

func ExampleNoNamespaces() {
	resetFlags()
	unmockAll := mockAll(
		[]string{"project-1"},
		[]*kubectl.Pod{
			&kubectl.Pod{Name: "pod-1", ContainerPorts: []int{1}, Containers: []string{"container-1"}},
		},
		[]*gcloud.Cluster{
			&gcloud.Cluster{Name: "cluster-1", Location: "location-1"},
		},
		"POD",
		[]string{},
	)
	defer unmockAll()
	os.Args = []string{"goproxie", "-local_port=1234"}
	main()
	// Output:
	// GCP Project: project-1
	// Cluster: cluster-1
	// Could not find any GCP Clusters
}

func ExampleNoPods() {
	resetFlags()
	unmockAll := mockAll(
		[]string{"project-1"},
		[]*kubectl.Pod{},
		[]*gcloud.Cluster{
			&gcloud.Cluster{Name: "cluster-1", Location: "location-1"},
		},
		"POD",
		[]string{"namespace-1"},
	)
	defer unmockAll()
	os.Args = []string{"goproxie", "-local_port=1234"}
	main()
	// Output:
	// GCP Project: project-1
	// Cluster: cluster-1
	// K8S Namespace: namespace-1
	// Could not find any K8S Pods in namespace namespace-1
}
