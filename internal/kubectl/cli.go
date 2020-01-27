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
	Name          string
	Namespace     string
	ContainerPort int
}

func NamespacesList() []string {
	out, err := exec.Command(kubectlPath, "get", "namespaces", "-o", "name").Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Fields(string(out))
}

func PodsList() []*Pod {
	out, err := exec.Command(kubectlPath, "get", "pods", "-o=custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,PORT:.spec.containers[0].ports[0].containerPort", "--all-namespaces=true").Output()
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
		port, _ := strconv.Atoi(tokens[2])
		pods = append(pods, &Pod{Name: tokens[0], Namespace: tokens[1], ContainerPort: port})
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
