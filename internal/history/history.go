package history

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/AckeeCZ/goproxie/internal/gcloud"
	"github.com/AckeeCZ/goproxie/internal/kubectl"
	"github.com/AckeeCZ/goproxie/internal/store"
	"github.com/AlecAivazis/survey/v2"
)

const KeyCommands = "history.commands"

func extractPodBaseName(pod *kubectl.Pod) string {
	// TODO
	return pod.Name
}

func StorePodProxy(projectId string, cluster *gcloud.Cluster, namespace string, pod *kubectl.Pod, localPort int, remotePort int) {

	record := fmt.Sprintf("-project=%v -cluster=%v -namespace=%v -pod=%v -local_port=%v -proxy_type=pod", projectId, cluster.Name, namespace, extractPodBaseName(pod), localPort)
	store.Append(KeyCommands, record)
}

func Browse() {
	storedCommands := store.Get(KeyCommands)
	commands := []string{}
	if storedCommands != nil {
		// 💡 Conversion problem: []interface{} -> []string
		// see https://stackoverflow.com/questions/44027826/convert-interface-to-string-in-golang
		for _, item := range storedCommands.([]interface{}) {
			commands = append(commands, fmt.Sprint(item))
		}
	}

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
