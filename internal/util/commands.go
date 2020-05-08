package util

import (
	"log"
	"os"
	"os/exec"
)

// RunCommand executes given command with args, automatically exits program on error.
func RunCommand(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

// RunSilentCommand is same as RunCommand but does not forward stderr.
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
