package yaml

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

// VisitorFunc allows functions with the required signature to implement
// ast.Visitor
type VisitorFunc func(node ast.Node) bool

func (v VisitorFunc) Visit(node ast.Node) ast.Visitor {
	if !v(node) {
		return nil
	}
	return v
}

// cmpPos provides a strict partial order on token.Position
func cmpPos(a *token.Position, b *token.Position) bool {
	if a.Line < b.Line {
		return true
	}
	if a.Line == b.Line {
		return a.Column < b.Column
	}
	return false
}

// FindNode locates the deepest AST node which contains the given line and
// column position.
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

// ParsedNode contains the YAML AST node and its unmarshalled value.
type ParsedNode struct {
	Node ast.Node

	// Value is the unmarshalled Value of `node`.
	Value interface{}
}

// PathFunc is the function handler to be whenever a node is located following
// a list of recursive YAML paths.
type PathFunc = func([]ParsedNode)

// VisitPath filters the given node by the list of YAML paths recursively,
// calling the `f` pathFunc whenever a node is found after visiting all paths
// in the list.
// For instance, if paths is {"$.spec.params[*]", "$.name"}, the `meta`
// function handler is called with the list {doc, param, name},
// in which `doc` is always the root document, `param` is one of the nodes
// in the sequence located at $.spec.params, and `name` is the `.name`
// property if `param`.
func VisitPath(node ast.Node, paths []*yaml.Path, f PathFunc, nodes ...ParsedNode) {
	pn := ParsedNode{
		Node: node,
	}
	// go-yaml bug:
	// ```yaml
	// val: >-
	//   name
	//   # if there is no trailing space here (or another node),
	//   # the value is unmarshalled as `nam`
	// ```
	// workaround: append a whitespace before unmarshalling
	err := yaml.Unmarshal([]byte(node.String()+" "), &pn.Value)
	if err != nil {
		return
	}
	nodes = append(nodes, pn)

	if len(paths) == 0 {
		f(nodes)
		return
	}

	p := paths[0]
	filtered, err := p.FilterNode(node)
	if err != nil || filtered == nil {
		return
	}

	if seq, ok := filtered.(*ast.SequenceNode); ok {
		for _, v := range seq.Values {
			if v == nil {
				continue
			}
			if _, isNull := v.(*ast.NullNode); isNull {
				continue
			}
			VisitPath(v, paths[1:], f, nodes...)
		}
	} else {
		VisitPath(filtered, paths[1:], f, nodes...)
	}
}

// TODO: move to yaml package
type VisitFunc = func(ast.Node)

// deprecated: should be substituded by visitPaths.
func VisitNodes(node ast.Node, depth int, f VisitFunc) {
	if node == nil {
		return
	}
	if _, isNull := node.(*ast.NullNode); isNull {
		return
	}
	if depth == 0 {
		f(node)
		return
	}
	for _, v := range node.(*ast.SequenceNode).Values {
		VisitNodes(v, depth-1, f)
	}
}
