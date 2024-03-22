package tekton

import (
	"testing"

	"github.com/cezarguimaraes/tekton-ls/internal/file"
)

func TestDocDiagnostics(t *testing.T) {
	_ = parseFile(file.TextDocument(string(pipelineDoc)))

	// f.Diagnostics()

	// f.Completions(protocol.Position{Line: 19, Character: 10})
}
