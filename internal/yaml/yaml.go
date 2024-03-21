package yaml

import (
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

var tst = []byte(`apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: hello
spec:
  params:
  - name: foo
    default: "hey"
  steps:
    - name: echo
      image: alpine
      test: ["a"]
      script: |
        #!/bin/sh
        echo "Hello $(pr"

# lua vim.lsp.start({ name = "tekton", cmd = { "go", "run", "main.go" }, root_dir = "." })
# lua vim.lsp.set_log_level 'debug'`)

type VisitorFunc func(node ast.Node) bool

func (v VisitorFunc) Visit(node ast.Node) ast.Visitor {
	if !v(node) {
		return nil
	}
	return v
}

func cmpPos(a *token.Position, b *token.Position) bool {
	if a.Line < b.Line {
		return true
	}
	if a.Line == b.Line {
		return a.Column < b.Column
	}
	return false
}

func FindNode(node ast.Node, line, col int) ast.Node {
	p := &token.Position{
		Line:   line,
		Column: col,
	}
	var res ast.Node
	// can be improved by culling the recursion
	ast.Walk(VisitorFunc(func(n ast.Node) bool {
		if _, ok := n.(*ast.NullNode); ok {
			// workaround for tentative go-yaml bug fix
			return false
		}
		tok := n.GetToken()
		nxt := tok.Next
		if nxt == nil {
			return true
		}

		// tok.Position <= p && p <= nxt.Position
		if !cmpPos(p, tok.Position) && !cmpPos(nxt.Position, p) {
			res = n
			// keep iterating to find the deepest node that contains the position
			return true
		}

		return true
	}), node)
	return res
}
