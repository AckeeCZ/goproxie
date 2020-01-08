package kubectl

import (
	"log"
	"os/exec"
	"strings"
)

func NamespacesList() []string {
	out, err := exec.Command("kubectl", "get", "namespaces", "-o", "name").Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Fields(string(out))
}
