package tekton

import (
	"os"
	"reflect"
	"testing"

	"github.com/cezarguimaraes/tekton-ls/internal/file"
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

func TestFileParseIdentifiers(t *testing.T) {
	f := ParseFile(file.File(string(multiDoc)))
	for docId := range 2 {
		for i, exp := range singleTCs {
			if i >= len(f.docs[docId].identifiers) {
				t.Fatalf("parseIdentifiers: got %d identifiers, want %d",
					len(f.docs[docId].identifiers),
					len(singleTCs),
				)
			}
			id := f.docs[docId].identifiers[i]
			if id.kind != exp.kind {
				t.Errorf("doc[%d].id[%d].kind: got %s, want %s", docId, i, id.kind, exp.kind)
			}
			if id.meta.Name() != exp.name {
				t.Errorf("doc[%d].id[%d].name: got %s, want %s", docId, i, id.meta.Name(), exp.name)
			}
			wantLine := exp.defLine + docId*30 // shift line values down for second doc
			if got := id.definition.GetToken().Position.Line; got != wantLine {
				t.Errorf("doc[%d].id[%d].definition.line: got %d, want %d", docId, i, got, wantLine)
			}
			if got := id.definition.GetToken().Position.Column; got != exp.defCol {
				t.Errorf("doc[%d].id[%d].definition.column: got %d, want %d", docId, i, got, exp.defCol)
			}

			var wantRefs []protocol.Range
			for _, ref := range exp.refs {
				wantRefs = append(wantRefs, protocol.Range{
					Start: protocol.Position{
						Line:      ref.Start.Line + uint32(30*docId),
						Character: ref.Start.Character,
					},
					End: protocol.Position{
						Line:      ref.End.Line + uint32(30*docId),
						Character: ref.End.Character,
					},
				})
			}
			gotRefs := wholeReferences(id)
			if !reflect.DeepEqual(gotRefs, wantRefs) {
				t.Errorf("doc[%d].id[%d].references:\ngot %v\nwant %v", docId, i, gotRefs, wantRefs)
			}
		}
	}
}

func TestFileFindReferences(t *testing.T) {
	f := ParseFile(file.File(string(multiDoc)))
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
					End:   protocol.Position{Line: 21, Character: 28},
				},
			},
		},
		{
			pos: protocol.Position{
				Line:      11,
				Character: 8,
			},
			refs: []protocol.Range{
				{
					Start: protocol.Position{Line: 24, Character: 20},
					End:   protocol.Position{Line: 24, Character: 33},
				},
			},
		},
		{
			pos: protocol.Position{
				Line:      36,
				Character: 10,
			},
			refs: []protocol.Range{
				{
					Start: protocol.Position{Line: 51, Character: 15},
					End:   protocol.Position{Line: 51, Character: 28},
				},
			},
		},
		{
			pos: protocol.Position{
				Line:      41,
				Character: 8,
			},
			refs: []protocol.Range{
				{
					Start: protocol.Position{Line: 54, Character: 20},
					End:   protocol.Position{Line: 54, Character: 33},
				},
			},
		},
	}
	for _, tc := range tcs {
		got := f.FindReferences(tc.pos)
		if !reflect.DeepEqual(got, tc.refs) {
			t.Errorf("FindReferences(%v):\ngot %v\nwant %v", tc.pos, got, tc.refs)
		}
	}
}
