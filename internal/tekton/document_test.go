package tekton

import (
	"os"
	"reflect"
	"testing"

	"github.com/cezarguimaraes/tekton-ls/internal/file"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var singleDoc, pipelineDoc []byte

func init() {
	singleDoc, _ = os.ReadFile("./testdata/single.yaml")
	pipelineDoc, _ = os.ReadFile("./testdata/pipeline.yaml")
}

type identTC struct {
	kind    identifierKind
	name    string
	defLine int // 1 based
	defCol  int // 1 based
	refs    []protocol.Range
}

var singleTCs = []identTC{
	{
		kind:    IdentKindParam,
		name:    "foo",
		defLine: 7,
		defCol:  11,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 21, Character: 15},
				End:   protocol.Position{Line: 21, Character: 28},
			},
		},
	},
	{
		kind:    IdentKindParam,
		name:    "b",
		defLine: 10,
		defCol:  11,
		refs:    nil,
	},
	{
		kind:    IdentKindParam,
		name:    "baz",
		defLine: 11,
		defCol:  11,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 24, Character: 20},
				End:   protocol.Position{Line: 24, Character: 33},
			},
		},
	},
	{
		kind:    IdentKindResult,
		name:    "foo",
		defLine: 14,
		defCol:  11,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 25, Character: 8},
				End:   protocol.Position{Line: 25, Character: 27},
			},
			{
				Start: protocol.Position{Line: 26, Character: 8},
				End:   protocol.Position{Line: 26, Character: 27},
			},
			{
				Start: protocol.Position{Line: 27, Character: 8},
				End:   protocol.Position{Line: 27, Character: 27},
			},
		},
	},
	{
		kind:    IdentKindWorkspace,
		name:    "test",
		defLine: 16,
		defCol:  11,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 28, Character: 8},
				End:   protocol.Position{Line: 28, Character: 31},
			},
		},
	},
	{
		kind:    IdentKindTask,
		name:    "hello",
		defLine: 4,
		defCol:  9,
	},
}

var pipeTCs = []identTC{
	{
		kind:    IdentKindWorkspace,
		name:    "source",
		defLine: 7,
		defCol:  13,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 16, Character: 21},
				End:   protocol.Position{Line: 16, Character: 27},
			},
			{
				Start: protocol.Position{Line: 24, Character: 21},
				End:   protocol.Position{Line: 24, Character: 27},
			},
		},
	},
	{
		kind:    IdentKindPipelineTask,
		name:    "gen-code",
		defLine: 9,
		defCol:  13,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 21, Character: 10},
				End:   protocol.Position{Line: 21, Character: 18},
			},
		},
	},
	{
		kind:    IdentKindPipelineTask,
		name:    "gen-code-2",
		defLine: 18,
		defCol:  13,
	},
}

var taskTCs = []identTC{
	{
		kind:    IdentKindParam,
		name:    "paramet",
		defLine: 33,
		defCol:  11,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 12, Character: 14},
				End:   protocol.Position{Line: 12, Character: 21},
			},
			{
				Start: protocol.Position{Line: 42, Character: 15},
				End:   protocol.Position{Line: 42, Character: 32},
			},
		},
	},
	{
		kind:    IdentKindResult,
		name:    "foo",
		defLine: 35,
		defCol:  11,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 45, Character: 20},
				End:   protocol.Position{Line: 45, Character: 39},
			},
		},
	},
	{
		kind:    IdentKindWorkspace,
		name:    "source",
		defLine: 37,
		defCol:  11,
	},
	{
		kind:    IdentKindTask,
		name:    "gen-code",
		defLine: 30,
		defCol:  9,
		refs: []protocol.Range{
			{
				Start: protocol.Position{Line: 10, Character: 14},
				End:   protocol.Position{Line: 10, Character: 22},
			},
			{
				Start: protocol.Position{Line: 19, Character: 14},
				End:   protocol.Position{Line: 19, Character: 22},
			},
		},
	},
}

func TestDocParseIdentifiers(t *testing.T) {
	single := ParseFile(file.File(string(singleDoc)))
	pipe := ParseFile(file.File(string(pipelineDoc)))

	tcs := []struct {
		name  string
		file  *File
		cases []identTC
		docId int
	}{
		{
			name:  "correctly parses identifiers of a single task",
			file:  single,
			cases: singleTCs,
		},
		{
			name:  "correctly parses identifiers of a pipeline",
			file:  pipe,
			cases: pipeTCs,
		},
		{
			name:  "correctly parses identifiers of a task",
			file:  pipe,
			cases: taskTCs,
			docId: 1,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.cases) != len(tc.file.docs[tc.docId].identifiers) {
				t.Errorf("parseIdentifiers: got %d identifiers, want %d",
					len(tc.file.docs[tc.docId].identifiers),
					len(tc.cases),
				)
			}

			for i, exp := range tc.cases {
				if i >= len(tc.file.docs[tc.docId].identifiers) {
					break
				}

				id := tc.file.docs[tc.docId].identifiers[i]
				if id.kind != exp.kind {
					t.Errorf("id[%d].kind: got %s, want %s", i, id.kind, exp.kind)
				}
				if id.meta.Name() != exp.name {
					t.Errorf("id[%d].name: got %s, want %s", i, id.meta.Name(), exp.name)
				}
				// TODO: test ident.prange instead of unreliable token
				if got := id.definition.GetToken().Position.Line; got != exp.defLine {
					t.Errorf("id[%d].definition.line: got %d, want %d", i, got, exp.defLine)
				}
				if got := id.definition.GetToken().Position.Column; got != exp.defCol {
					t.Errorf("id[%d].definition.column: got %d, want %d", i, got, exp.defCol)
				}
				gotRefs := locationToRange(wholeReferences(id))
				if !reflect.DeepEqual(gotRefs, exp.refs) {
					t.Errorf("id[%d].references:\ngot %v\nwant %v", i, gotRefs, exp.refs)
				}
			}
		})
	}
}

func TestDocFindReferences(t *testing.T) {
	f := ParseFile(file.File(string(singleDoc)))
	p := ParseFile(file.File(string(pipelineDoc)))

	single := f.docs[0]
	pipe := p.docs[0]

	tcs := []struct {
		doc  *Document
		pos  protocol.Position
		refs []protocol.Range
	}{
		{
			doc: single,
			pos: protocol.Position{
				Line:      6,
				Character: 12,
			},
			refs: []protocol.Range{
				{
					Start: protocol.Position{Line: 21, Character: 15},
					End:   protocol.Position{Line: 21, Character: 28},
				},
			},
		},
		{
			doc: single,
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
			doc: pipe,
			pos: protocol.Position{
				Line:      6,
				Character: 12,
			},
			refs: []protocol.Range{
				{
					Start: protocol.Position{Line: 16, Character: 21},
					End:   protocol.Position{Line: 16, Character: 27},
				},
				{
					Start: protocol.Position{Line: 24, Character: 21},
					End:   protocol.Position{Line: 24, Character: 27},
				},
			},
		},
	}
	for _, tc := range tcs {
		got := locationToRange(tc.doc.findReferences(tc.pos))
		if !reflect.DeepEqual(got, tc.refs) {
			t.Errorf("FindReferences(%v):\ngot %v\nwant %v", tc.pos, got, tc.refs)
		}
	}
}

func TestDocParseIdentifiers_Cases(t *testing.T) {
	tc := []struct {
		contents string
	}{
		{
			contents: `apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: task_a
spec:
  steps:
    - name: echo
      script: ["/bin/sh", "-c", "echo Hello World"]
`,
		},
	}
	for _, tc := range tc {
		_ = ParseFile(file.File(tc.contents))
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
