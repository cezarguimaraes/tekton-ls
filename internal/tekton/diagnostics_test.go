package tekton

import (
	"testing"

	"github.com/cezarguimaraes/tekton-ls/internal/file"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestDocDiagnostics(t *testing.T) {
	f := ParseFile(file.File(string(pipelineDoc)))

	f.Diagnostics()

	f.Completions(protocol.Position{Line: 19, Character: 10})
}
