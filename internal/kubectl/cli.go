package kubectl

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/AckeeCZ/goproxie/internal/util"
)

var kubectlPath = "kubectl"

var runCommand = util.RunCommand

// SetKubectlPath sets the executable path to kubectl bin.
func SetKubectlPath(path string) {
	kubectlPath = path
}

// Pod structure
type Pod struct {
	Name           string
	Containers     []string
	ContainerPorts []int
	AppLabel       string
}

// NamespacesList returns the list of k8s namespaces
func NamespacesList() []string {
	return strings.Fields(runCommand(kubectlPath, "get", "namespaces", "-o=custom-columns=NAME:.metadata.name", "--no-headers"))
}

// PodsList returns the list of k8s pods from the given namespace
func PodsList(namespace string) []*Pod {
	out := runCommand(kubectlPath, "get", "pods", "--namespace", namespace, "--no-headers",
		"-o=custom-columns=NAME:.metadata.name,CONTAINERS:spec.containers[*].name,PORTS:.spec.containers[*].ports[*].containerPort,LABELS=:.metadata.labels.app")
	lines := strings.Split(out, "\n")
	pods := []*Pod{}
	for _, line := range lines {
		tokens := strings.Fields(line)
		if len(tokens) < 3 {
			continue
		}
		containers := strings.Split(tokens[1], ",")
		portsStr := strings.Split(tokens[2], ",")
		ports := make([]int, 0, len(portsStr))
		for _, portStr := range portsStr {
			port, err := strconv.Atoi(portStr)
			if (err == nil) {
				ports = append(ports, port)
			}
		}
		pods = append(pods, &Pod{Name: tokens[0], Containers: containers, ContainerPorts: ports, AppLabel: tokens[3]})
	}
	return pods
}

// PortForward executes kubectl's 'port-forward'.
// Local port is bound to 0.0.0.0. via '--address'.
func PortForward(podID string, localPort int, remotePort int, namespace string) {
	cmd := exec.Command(kubectlPath, "port-forward", podID, fmt.Sprintf("%v:%v", localPort, remotePort), "--namespace", namespace, "--address", "0.0.0.0")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
