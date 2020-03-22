package history

import (
	"testing"

	"github.com/AckeeCZ/goproxie/internal/kubectl"
)

func TestExtractPodBaseName(t *testing.T) {
	t.Skip()
	result := extractPodBaseName(&kubectl.Pod{Name: "dashed-app-name-598bf9b49d-9h2j7"})
	if result != "dashed-app-name" {
		t.Errorf("Expected dashed-app-name-598bf9b49d-9h2j7 to convert into dashed-app-name, got %v", result)
	}
}
