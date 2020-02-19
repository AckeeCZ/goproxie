package kubectl

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var kubectlPath = "kubectl"

func SetKubectlPath(path string) {
	kubectlPath = path
}

type Pod struct {
	Name           string
	Namespace      string
	Containers     []string
	ContainerPorts []int
}

func NamespacesList() []string {
	out, err := exec.Command(kubectlPath, "get", "namespaces", "-o", "name").Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Fields(string(out))
}

func PodsList() []*Pod {
	out, err := exec.Command(kubectlPath, "get", "pods", "-o=custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,CONTAINERS:spec.containers[*].name,PORTS:.spec.containers[*].ports[*].containerPort", "--all-namespaces=true").Output()
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(out), "\n")
	pods := []*Pod{}
	for _, line := range lines {
		tokens := strings.Fields(line)
		if len(tokens) < 3 {
			continue
		}
		containers := strings.Split(tokens[2], ",")
		portsStr := strings.Split(tokens[3], ",")
		ports := make([]int, 0, len(portsStr))
		for _, portStr := range portsStr {
			port, _ := strconv.Atoi(portStr)
			ports = append(ports, port)
		}
		pods = append(pods, &Pod{Name: tokens[0], Namespace: tokens[1], Containers: containers, ContainerPorts: ports})
	}
	return pods
}

func PortForward(podId string, localPort int, remotePort int, namespace string) {
	cmd := exec.Command(kubectlPath, "port-forward", podId, fmt.Sprintf("%v:%v", localPort, remotePort), "--namespace", namespace)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
