package sqlproxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
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

// CloudSQLInstanceType is one of supported POSTGRES, MYSQL, SQLSERVER
type CloudSQLInstanceType string

const (
	// TypePostgres cloud SQL instance
	TypePostgres CloudSQLInstanceType = "POSTGRES"
	// TypeMySQL cloud SQL instance
	TypeMySQL CloudSQLInstanceType = "MYSQL"
	// TypeSQLServer cloud SQL instance
	TypeSQLServer CloudSQLInstanceType = "SQLSERVER"
	// TypeUnknown for unknown cloud SQL instance type
	TypeUnknown CloudSQLInstanceType = "UNKNOWN"
)

// CloudSQLInstance struct
type CloudSQLInstance struct {
	// project[:region]:instance-name
	ConnectionName string
	Type           CloudSQLInstanceType
	DefaultPort    int
}

// GetDefaultPortForType returns default port for given database type
func GetDefaultPortForType(dbType CloudSQLInstanceType) int {
	switch dbType {
	case TypePostgres:
		return 5432
	case TypeMySQL:
		return 3306
	case TypeSQLServer:
		return 1433
	default:
		return 0
	}
}

func getSQLInstanceType(in *sqladmin.DatabaseInstance) CloudSQLInstanceType {
	if strings.Contains(in.DatabaseVersion, "POSTGRES") {
		return TypePostgres
	}
	if strings.Contains(in.DatabaseVersion, "SQLSERVER") {
		return TypeSQLServer
	}
	if strings.Contains(in.DatabaseVersion, "MYSQL") {
		return TypeMySQL
	}
	return TypeUnknown
}

// GetInstancesList gets list of Cloud SQL instances for given projects
func GetInstancesList(projects []string) ([]CloudSQLInstance, error) {
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

	ch := make(chan CloudSQLInstance)
	var wg sync.WaitGroup
	wg.Add(len(projects))
	for _, proj := range projects {
		proj := proj
		go func() {
			err := sql.Instances.List(proj).Pages(ctx, func(r *sqladmin.InstancesListResponse) error {
				for _, in := range r.Items {
					// The Proxy is only support on Second Gen
					if in.BackendType == "SECOND_GEN" {
						connName := fmt.Sprintf("%s:%s:%s", in.Project, in.Region, in.Name)
						dbType := getSQLInstanceType(in)
						ch <- CloudSQLInstance{ConnectionName: connName, Type: dbType, DefaultPort: GetDefaultPortForType(dbType)}
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
	var ret []CloudSQLInstance
	for x := range ch {
		ret = append(ret, x)
	}
	if len(ret) == 0 {
		return nil, fmt.Errorf("no Cloud SQL Instances found in these projects: %v", projects)
	}
	return ret, nil
}
