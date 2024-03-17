package tekton

import (
	"os"
	"testing"

	"github.com/cezarguimaraes/tekton-lsp/internal/file"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var multiDoc []byte

func init() {
	multiDoc, _ = os.ReadFile("./testdata/multi.yaml")
}

func TestFindDoc(t *testing.T) {
	file := ParseFile(file.File(string(multiDoc)))

	file.findDoc(protocol.Position{Line: 0, Character: 0})
}
