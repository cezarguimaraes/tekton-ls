package tekton

import (
	"testing"

	"github.com/cezarguimaraes/tekton-lsp/internal/file"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var document = `apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: hello
spec:
  parameters:
  - name: foo
    description: "my param foo meant for stuff"
    default: "hey"
  - name: b
  - name: >-
      baz
  results:
  - name: foo
  workspaces:
  - name: test
  steps:
    - name: echo
      image: idk
      test: "foo"
      script: |
        #!/bin/sh
        echo "Hello $(params.baz)
        $(results.foo.path)
        $(results.foo.path)
        $(results.foo.path)
        $(workspaces.test.path)`

func TestParseIdentifiers(t *testing.T) {
	f := ParseFile(file.File(document))
	ids := f.parseIdentifiers()
	for _, id := range ids {
		t.Log(id.kind, id.meta.Name(), id.meta.Documentation(), id.definition.GetToken().Position, id.definition.String())
	}
}

func TestFindReference(t *testing.T) {
	f := ParseFile(file.File(document))
	pos := protocol.Position{
		Line:      25,
		Character: 20,
	}
	ref := f.findReference(pos)
	if ref == nil {
		t.Fatalf("reference not found")
	}
	t.Logf("found ident %s %s", ref.ident.kind, ref.ident.meta.Name())

	def := f.Definition(pos)
	t.Logf("found definition at %d %d", def.Line, def.Character)
}
