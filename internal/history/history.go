package history

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/AckeeCZ/goproxie/internal/fsconfig"
	"github.com/AckeeCZ/goproxie/internal/gcloud"
	"github.com/AckeeCZ/goproxie/internal/kubectl"
	"github.com/AckeeCZ/goproxie/internal/sqlproxy"
	"github.com/AlecAivazis/survey/v2"
)

// KeyCommands defines the configration key
const KeyCommands = "history.commands"

// StorePodProxy appends the given run configuration to history commands
// in a form of non-interactive goproxie arguments.
func StorePodProxy(projectID string, cluster *gcloud.Cluster, namespace string, pod *kubectl.Pod, localPort int, remotePort int) {

	record := fmt.Sprintf("-project=%v -cluster=%v -namespace=%v -pod=%v -local_port=%v -proxy_type=pod", projectID, cluster.Name, namespace, pod.AppLabel, localPort)
	fsconfig.AppendHistoryCommand(record)
}

// StoreCloudSQLProxy appends the given run configuration to history commands
func StoreCloudSQLProxy(projectID string, instance sqlproxy.CloudSQLInstance, localPort int) {
	record := fmt.Sprintf("-project=%v -sql_instance=%v -local_port=%v -proxy_type=sql", projectID, instance.ConnectionName, localPort)
	fsconfig.AppendHistoryCommand(record)
}

func ListRaw() []string {
	return fsconfig.GetConfig().History.Commands
}

type Item struct {
	ProjectID   string `json:"projectId"`
	Cluster     string `json:"cluster"`
	SqlInstance string `json:"sqlInstance"`
	LocalPort   int    `json:"localPort"`
	RemotePort  int    `json:"remotePort"`
	ProxyType   string `json:"proxyType"`
	Pod         string `json:"pod"`
	Namespace   string `json:"namespace"`
}

func List() []Item {
	stringCommands := ListRaw()

	items := []Item{}
	for _, raw := range stringCommands {
		argsTokens := strings.Fields(raw)
		item := Item{}
		for _, token := range argsTokens {
			argTokens := strings.Split(token, "=")
			flag := argTokens[0]
			value := argTokens[1]
			switch flag {
			case "-project":
				item.ProjectID = value
			case "-sql_instance":
				item.SqlInstance = value
			case "-proxy_type":
				item.ProxyType = value
			case "-local_port":
				item.LocalPort, _ = strconv.Atoi(value)
			case "-remote_port":
				item.RemotePort, _ = strconv.Atoi(value)
			case "-pod":
				item.Pod = value
			case "-namespace":
				item.Namespace = value
			case "-cluster":
				item.Cluster = value

			}
		}
		items = append(items, item)
	}
	return items
}

// Browse lets user choose from stored commands.
// Goproxie is executed with given arguments.
func Browse() {
	commands := ListRaw()

	if len(commands) == 0 {
		fmt.Println("History is empty")
		os.Exit(0)
	}

	pickedCommand := ""
	survey.AskOne(&survey.Select{
		Message: "Pick command from history",
		Options: commands,
	}, &pickedCommand)
	proxieBin := os.Args[0]
	cmd := exec.Command(proxieBin, append(strings.Fields(pickedCommand), "--no-save")...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
