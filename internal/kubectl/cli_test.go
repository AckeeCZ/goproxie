package kubectl

import "testing"

// Exact command results

// Should contain <none> ports and multi-ports
var mockPodsList = `acme-rockets-v0.3.0-74bf544f8b-lzc5b                      event-exporter,prometheus-to-sd-exporter      <none>
metrics-server-v0.3.3-6d96fcc55-2qtm8                       metrics-server,metrics-server-nanny           443
traefik-ig-7646cb565d-9zxv6                                 traefik                                       80,443,8080,8081
`
var mockNamespacesList = `acme-sro-development
default
infrastructure-development
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

func TestNamespacesList(t *testing.T) {
	unmock := mockRunCommand(mockNamespacesList)
	defer unmock()
	result := NamespacesList()
	expectedItems := []string{
		"acme-sro-development",
		"default",
		"infrastructure-development",
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

func TestPodsList(t *testing.T) {
	unmock := mockRunCommand(mockPodsList)
	defer unmock()
	result := PodsList("anynamespace")
	expectedItems := []*Pod{
		&Pod{
			Name: "acme-rockets-v0.3.0-74bf544f8b-lzc5b",
			// TODO Is it OK to have [0] for <none>? Because it does that now.
			ContainerPorts: []int{0},
			Containers: []string{
				"event-exporter",
				"prometheus-to-sd-exporter",
			},
		},
		&Pod{
			Name:           "metrics-server-v0.3.3-6d96fcc55-2qtm8",
			ContainerPorts: []int{443},
			Containers: []string{
				"metrics-server",
				"metrics-server-nanny",
			},
		},
		&Pod{
			Name:           "traefik-ig-7646cb565d-9zxv6",
			ContainerPorts: []int{80, 443, 8080, 8081},
			Containers: []string{
				"traefik",
			},
		},
	}
	for i, expectedItem := range expectedItems {
		resultItem := result[i]
		if expectedItem.Name != resultItem.Name {
			t.Errorf("Expected `%v` does not match result `%v`", expectedItem, resultItem)
		}
		for i, expectedPort := range expectedItem.ContainerPorts {
			resultPort := resultItem.ContainerPorts[i]
			if expectedPort != resultPort {
				t.Errorf("Expected `%v` does not match result `%v`", expectedPort, resultPort)
			}
		}
		if len(expectedItem.ContainerPorts) != len(resultItem.ContainerPorts) {
			t.Errorf("Expected len `%v` does not match result `%v`", expectedItem.ContainerPorts, resultItem.ContainerPorts)
		}
		for i, expectedContainer := range expectedItem.Containers {
			resultContainer := resultItem.Containers[i]
			if expectedContainer != resultContainer {
				t.Errorf("Expected `%v` does not match result `%v`", expectedContainer, resultContainer)
			}
		}
		if len(expectedItem.Containers) != len(resultItem.Containers) {
			t.Errorf("Expected len `%v` does not match result `%v`", len(expectedItem.Containers), len(resultItem.Containers))
		}
	}
	if len(expectedItems) != len(result) {
		t.Errorf("Expected len `%v` does not match result `%v`", len(expectedItems), len(result))
	}
}
