package gcloud

import (
	"testing"
)

// Exact command results
var mockProjectsList = `acme-sro-development
snackee
infrastructure-1188
`
var mockClustersList = `production	europe-west1-d
`

func mockRunCommand(mockResponse string) func() {
	originalRunCommand := runCommand
	runCommand = func(cmd string, args ...string) string {
		return mockResponse
	}
	return func() {
		runCommand = originalRunCommand
	}
}

func TestProjectsList(t *testing.T) {
	t.Errorf("Fail test")
	unmock := mockRunCommand(mockProjectsList)
	defer unmock()
	result := ProjectsList()
	expectedItems := []string{
		"acme-sro-development",
		"snackee",
		"infrastructure-1188",
	}
	if len(expectedItems) != len(result) {
		t.Errorf("Expected len `%v` does not match result `%v`", len(expectedItems), len(result))
	}
	for i, line := range expectedItems {
		if line != result[i] {
			t.Errorf("Expected `%v` does not match result `%v`", line, result[i])
		}
	}
}

func TestContainerClustersList(t *testing.T) {
	unmock := mockRunCommand(mockClustersList)
	defer unmock()
	result := ContainerClustersList("anyproject")
	expectedItems := []*Cluster{
		&Cluster{
			Name:     "production",
			Location: "europe-west1-d",
		},
	}
	if len(expectedItems) != len(result) {
		t.Errorf("Expected len `%v` does not match result `%v`", len(expectedItems), len(result))
	}
	for i, item := range expectedItems {
		if item.Name != result[i].Name || item.Location != result[i].Location {
			t.Errorf("Expected `%v` does not match result `%v`", item, result[i])
		}
	}
}
