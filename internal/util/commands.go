package util

import (
	"log"
	"os"
	"os/exec"
)

func RunCommand(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

// TODO: Remove and refactor RunCommand to print stderr only when err happens
// due to gcloud printing to stderr it's debug messages
func RunSilentCommand(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}
