package gcloud

import (
	"strings"

	"github.com/AckeeCZ/goproxie/internal/util"
)

var gcloudPath = "gcloud"

var runCommand = util.RunCommand

// SetGcloudPath sets the executable path to gcloud bin.
func SetGcloudPath(path string) {
	gcloudPath = path
}

// ProjectsList returns the list of google cloud projects
func ProjectsList() []string {
	return strings.Fields(runCommand(gcloudPath, "projects", "list", "--format", "value(projectId)"))
}

// Cluster structure
type Cluster struct {
	Name     string
	Location string
}

//ContainerClustersList returns the list of GCP clusters
func ContainerClustersList(projectID string) []*Cluster {
	out := runCommand(gcloudPath, "container", "clusters", "list", "--format", "value(name, location)", "--project", projectID)
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

// GetClusterCredentials gets credentials for the given GCP cluster
func GetClusterCredentials(projectID string, cluster *Cluster) {
	util.RunSilentCommand(gcloudPath, "container", "clusters", "get-credentials", cluster.Name, "--project", projectID, "--zone", cluster.Location)
}
