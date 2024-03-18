package tekton

import (
	"os"
	"reflect"
	"testing"

	"github.com/cezarguimaraes/tekton-lsp/internal/file"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var singleDoc []byte

func init() {
	singleDoc, _ = os.ReadFile("./testdata/single.yaml")
}

var singleTCs = []struct {
	kind    identifierKind
	name    string
	defLine int // 1 based
	defCol  int // 1 based
	refs    []protocol.Range
}{
	{
		kind:    IdentParam,
		name:    "foo",
		defLine: 7,
		defCol:  11,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 21, Character: 15},
				End:   protocol.Position{Line: 21, Character: 27},
			},
		},
	},
	{
		kind:    IdentParam,
		name:    "b",
		defLine: 10,
		defCol:  11,
		refs:    nil,
	},
	{
		kind:    IdentParam,
		name:    "baz",
		defLine: 11,
		defCol:  11,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 24, Character: 20},
				End:   protocol.Position{Line: 24, Character: 32},
			},
		},
	},
	{
		kind:    IdentResult,
		name:    "foo",
		defLine: 14,
		defCol:  11,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 25, Character: 8},
				End:   protocol.Position{Line: 25, Character: 26},
			},
			{
				Start: protocol.Position{Line: 26, Character: 8},
				End:   protocol.Position{Line: 26, Character: 26},
			},
			{
				Start: protocol.Position{Line: 27, Character: 8},
				End:   protocol.Position{Line: 27, Character: 26},
			},
		},
	},
	{
		kind:    IdentWorkspace,
		name:    "test",
		defLine: 16,
		defCol:  11,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 28, Character: 8},
				End:   protocol.Position{Line: 28, Character: 30},
			},
		},
	},
}

func TestDocParseIdentifiers(t *testing.T) {
	single := ParseFile(file.File(string(singleDoc)))
	for i, exp := range singleTCs {
		if i >= len(single.docs[0].identifiers) {
			t.Fatalf("parseIdentifiers: got %d identifiers, want %d",
				len(single.docs[0].identifiers),
				len(singleTCs),
			)
		}
		id := single.docs[0].identifiers[i]
		if id.kind != exp.kind {
			t.Errorf("id[%d].kind: got %s, want %s", i, id.kind, exp.kind)
		}
		if id.meta.Name() != exp.name {
			t.Errorf("id[%d].name: got %s, want %s", i, id.meta.Name(), exp.name)
		}
		if got := id.definition.GetToken().Position.Line; got != exp.defLine {
			t.Errorf("id[%d].definition.line: got %d, want %d", i, got, exp.defLine)
		}
		if got := id.definition.GetToken().Position.Column; got != exp.defCol {
			t.Errorf("id[%d].definition.column: got %d, want %d", i, got, exp.defCol)
		}
		if !reflect.DeepEqual(id.references, exp.refs) {
			t.Errorf("id[%d].references:\ngot %v\nwant %v", i, id.references, exp.refs)
		}
	}
}

func TestDocFindReferences(t *testing.T) {
	f := ParseFile(file.File(string(singleDoc)))
	tcs := []struct {
		pos  protocol.Position
		refs []protocol.Range
	}{
		{
			pos: protocol.Position{
				Line:      6,
				Character: 10,
			},
			refs: []protocol.Range{
				{
					Start: protocol.Position{Line: 21, Character: 15},
					End:   protocol.Position{Line: 21, Character: 27},
				},
			},
		},
		{
			pos: protocol.Position{
				Line:      11,
				Character: 9,
			},
			refs: []protocol.Range{
				{
					Start: protocol.Position{Line: 24, Character: 20},
					End:   protocol.Position{Line: 24, Character: 32},
				},
			},
		},
	}
	for _, tc := range tcs {
		got := f.docs[0].findReferences(tc.pos)
		if !reflect.DeepEqual(got, tc.refs) {
			t.Errorf("FindReferences(%v):\ngot %v\nwant %v", tc.pos, got, tc.refs)
		}
	}
}

func TestDocFindDefinition(t *testing.T) {
	f := ParseFile(file.File(string(singleDoc)))
	pos := protocol.Position{
		Line:      25,
		Character: 20,
	}
	ref := f.docs[0].findDefinition(pos)
	if ref == nil {
		t.Fatalf("reference not found")
	}
	t.Logf("found ident %s %s", ref.ident.kind, ref.ident.meta.Name())

	def := f.Definition(pos)
	t.Logf("found definition at %v", def)
}
