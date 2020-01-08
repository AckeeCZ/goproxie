package gcloud

import (
	"log"
	"os/exec"
	"strings"
)

func ProjectsList() []string {
	out, err := exec.Command("gcloud", "projects", "list", "--format", "value(projectId)").Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Fields(string(out))
}

func ContainerClustersList() []string {
	out, err := exec.Command("gcloud", "container", "clusters", "list", "--format", "value(name)").Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Fields(string(out))
}
