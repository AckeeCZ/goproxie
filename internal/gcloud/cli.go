package gcloud

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

var gcloudPath = "gcloud"

func SetGcloudPath(path string) {
	gcloudPath = path
}

func ProjectsList() []string {
	out, err := exec.Command(gcloudPath, "projects", "list", "--format", "value(projectId)").Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Fields(string(out))
}

type Cluster struct {
	Name     string
	Location string
}

func ContainerClustersList(projectId string) []*Cluster {
	out, err := exec.Command(gcloudPath, "container", "clusters", "list", "--format", "value(name, location)", "--project", projectId).Output()
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(out), "\n")
	clusters := []*Cluster{}
	for _, line := range lines {
		split := strings.Fields(line)
		if len(split) >= 2 {
			clusters = append(clusters, &Cluster{Name: split[0], Location: split[1]})
		}
	}
	return clusters
	// return Cluster{name: results[0], location: results[1]}
}

func GetClusterCredentials(projectID string, cluster *Cluster) {
	cmd := exec.Command(gcloudPath, "container", "clusters", "get-credentials", cluster.Name, "--project", projectID, "--zone", cluster.Location)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
