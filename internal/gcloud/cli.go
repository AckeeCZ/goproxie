package gcloud

import (
	"strings"

	"github.com/AckeeCZ/goproxie/internal/util"
)

var gcloudPath = "gcloud"

var runCommand = util.RunCommand

func SetGcloudPath(path string) {
	gcloudPath = path
}

func ProjectsList() []string {
	return strings.Fields(runCommand(gcloudPath, "projects", "list", "--format", "value(projectId)"))
}

type Cluster struct {
	Name     string
	Location string
}

func ContainerClustersList(projectId string) []*Cluster {
	out := runCommand(gcloudPath, "container", "clusters", "list", "--format", "value(name, location)", "--project", projectId)
	lines := strings.Split(out, "\n")
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
	util.RunSilentCommand(gcloudPath, "container", "clusters", "get-credentials", cluster.Name, "--project", projectID, "--zone", cluster.Location)
}
