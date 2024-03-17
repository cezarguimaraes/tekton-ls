package tekton

import (
	"testing"

	"github.com/cezarguimaraes/tekton-lsp/internal/file"
)

var document = `apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: hello
spec:
  parameters:
  - name: "foo"
    description: "my param foo meant for stuff"
    default: "hey"
  - name: b
  - name: >-
      middle
    t: test
  - name: >-
      end
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
        $(workspaces.test.path)
        $(params.baz)
        $(params.foo)
        $(results.foo.path)
        $(workspaces.test.path)
`

func TestParseIdentifiers(t *testing.T) {
	f := ParseFile(file.File(document))
	ids := f.parseIdentifiers()
	for _, id := range ids {
		t.Log(id.kind, id.meta.Name(), id.meta.Documentation(), id.definition.GetToken().Position, id.definition.String())
	}
}
