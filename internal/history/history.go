package history

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/AckeeCZ/goproxie/internal/fsconfig"
	"github.com/AckeeCZ/goproxie/internal/gcloud"
	"github.com/AckeeCZ/goproxie/internal/kubectl"
	"github.com/AckeeCZ/goproxie/internal/sqlproxy"
	"github.com/AlecAivazis/survey/v2"
	"golang.org/x/net/context"
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
	ID          string
	ProjectID   string
	Cluster     string
	SqlInstance string
	LocalPort   int
	RemotePort  int
	ProxyType   string
	Pod         string
	Namespace   string
	Raw         string
}

func ParseRaw(raw string) *Item {
	argsTokens := strings.Fields(raw)
	item := Item{}
	item.ID = strings.ReplaceAll(raw, " ", "")
	item.Raw = raw
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
	return &item
}

func List() []Item {
	stringCommands := ListRaw()

	items := []Item{}
	for _, raw := range stringCommands {
		item := ParseRaw(raw)
		items = append(items, *item)
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
	ExecHistoryItem(pickedCommand)
}

type SpawnedHistoryCommand struct {
	Raw  string
	Kill func()
	Wait func()
}

func ExecHistoryItem(raw string) *SpawnedHistoryCommand {
	proxieBin := os.Args[0]
	// 💡 A good example of contexts. This one adds a cancel function,
	// so for calls that supports this (in this case CommandContext, just a Command
	// but allows you to pass in context).
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, proxieBin, append(strings.Fields(raw), "--no-save")...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	// Kill subprocess if the main process gets killed
	// https://groups.google.com/g/golang-nuts/c/XoQ3RhFBJl8
	// Without this I felt that _sometimes_ the opened connections kept being open
	// occupying the port.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}
	spawn := &SpawnedHistoryCommand{}
	life := sync.WaitGroup{}
	spawn.Kill = cancel
	spawn.Wait = func() {
		life.Wait()
	}
	life.Add(1)
	go func() {
		err := cmd.Run()
		if err != nil {
			log.Print(err)
		}
		life.Done()
	}()
	return spawn
}
