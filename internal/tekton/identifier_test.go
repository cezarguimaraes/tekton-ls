package tekton

import (
	"os"
	"testing"

	"github.com/cezarguimaraes/tekton-lsp/internal/file"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var (
	singleDoc []byte
	multiDoc  []byte
)

func init() {
	singleDoc, _ = os.ReadFile("./testdata/single.yaml")
	multiDoc, _ = os.ReadFile("./testdata/multiDoc.yaml")
}

func TestParseIdentifiers(t *testing.T) {
	f := ParseFile(file.File(string(singleDoc)))
	expected := []struct {
		kind    identifierKind
		name    string
		defLine int // 1 based
		defCol  int // 1 based
	}{
		{
			kind:    IdentParam,
			name:    "foo",
			defLine: 7,
			defCol:  11,
		},
		{
			kind:    IdentParam,
			name:    "b",
			defLine: 10,
			defCol:  11,
		},
		{
			kind:    IdentParam,
			name:    "baz",
			defLine: 11,
			defCol:  11,
		},
		{
			kind:    IdentResult,
			name:    "foo",
			defLine: 14,
			defCol:  11,
		},
		{
			kind:    IdentWorkspace,
			name:    "test",
			defLine: 16,
			defCol:  11,
		},
	}
	for i, id := range f.identifiers {
		exp := expected[i]
		if id.kind != exp.kind {
			t.Errorf("id[%d].kind: got %s, want %s", i, id.kind, exp.kind)
		}
		t.Log(id.references)
	}
}

func TestFindReferences(t *testing.T) {
	f := ParseFile(file.File(string(multiDoc)))
	tcs := []struct {
		pos  protocol.Position
		refs [][]int
	}{
		{
			pos: protocol.Position{
				Line:      6,
				Character: 10,
			},
			refs: [][]int{
				{},
			},
		},
	}
	for _, tc := range tcs {
		_ = f
		_ = tc
	}
}

func TestFindDefinition(t *testing.T) {
	f := ParseFile(file.File(string(singleDoc)))
	pos := protocol.Position{
		Line:      25,
		Character: 20,
	}
	ref := f.findDefinition(pos)
	if ref == nil {
		t.Fatalf("reference not found")
	}
	t.Logf("found ident %s %s", ref.ident.kind, ref.ident.meta.Name())

	def := f.Definition(pos)
	t.Logf("found definition at %d %d", def.Line, def.Character)
}
