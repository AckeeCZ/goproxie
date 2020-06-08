package sqlproxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/GoogleCloudPlatform/cloudsql-proxy/logging"
	"github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/proxy"
	"golang.org/x/oauth2/google"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

const (
	host = "https://sqladmin.googleapis.com"
)

// CreateHTTPAuthClient creats http auth client for google apis
func CreateHTTPAuthClient() *http.Client {
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, proxy.SQLScope)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// CloudSQLInstance in form project[:region]:instance-name
type CloudSQLInstance = string

// GetInstancesList gets list of Cloud SQL instances for given projects
func GetInstancesList(projects []string) ([]string, error) {
	ctx := context.Background()
	client := CreateHTTPAuthClient()
	if len(projects) == 0 {
		// No projects requested.
		return nil, nil
	}

	sql, err := sqladmin.New(client)
	if err != nil {
		return nil, err
	}
	if host != "" {
		sql.BasePath = host
	}

	ch := make(chan string)
	var wg sync.WaitGroup
	wg.Add(len(projects))
	for _, proj := range projects {
		proj := proj
		go func() {
			err := sql.Instances.List(proj).Pages(ctx, func(r *sqladmin.InstancesListResponse) error {
				for _, in := range r.Items {
					// The Proxy is only support on Second Gen
					if in.BackendType == "SECOND_GEN" {
						ch <- fmt.Sprintf("%s:%s:%s", in.Project, in.Region, in.Name)
					}
				}
				return nil
			})
			if err != nil {
				logging.Errorf("Error listing instances in %v: %v", proj, err)
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	var ret []string
	for x := range ch {
		ret = append(ret, x)
	}
	if len(ret) == 0 {
		return nil, fmt.Errorf("no Cloud SQL Instances found in these projects: %v", projects)
	}
	return ret, nil
}
