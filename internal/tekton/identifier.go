package tekton

import (
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type identifierKind int

const (
	IdentKindParam identifierKind = iota
	IdentKindResult
	IdentKindWorkspace
	IdentKindPipelineTask
	IdentKindTask
)

func (k identifierKind) String() string {
	switch k {
	case IdentKindParam:
		return "parameter"
	case IdentKindResult:
		return "result"
	case IdentKindWorkspace:
		return "workspace"
	case IdentKindPipelineTask:
		return "pipelineTask"
	case IdentKindTask:
		return "task"
	}
	return ""
}

type identifier struct {
	kind       identifierKind
	meta       Meta
	definition ast.Node
	location   protocol.Location
	references [][]protocol.Location
}

type identLocator interface {
	matches(*identifier) bool
}

type kindNameLocator struct {
	kind identifierKind
	name string
}

func (l *kindNameLocator) matches(id *identifier) bool {
	return id.kind == l.kind && id.meta.Name() == l.name
}

type taskParamLocator struct {
	name     string
	taskName string
}

func (l *taskParamLocator) matches(id *identifier) bool {
	if id.kind != IdentKindParam {
		return false
	}
	if id.meta.Name() != l.name {
		return false
	}
	p, ok := id.meta.(*identParam)
	if !ok {
		return false
	}
	if p.parentKind != "task" {
		return false
	}
	if p.parentName != l.taskName {
		return false
	}
	return true
}

func (d *Document) getIdent(l identLocator) *identifier {
	for _, id := range d.identifiers {
		if !l.matches(id) {
			continue
		}
		return id
	}
	return nil
}

var identifiers = []struct {
	kind identifierKind
	meta func([]parsedNode) Meta

	paths []*yaml.Path
}{
	{
		kind: IdentKindParam,
		paths: []*yaml.Path{
			mustPathString("$.spec.params[*]"),
			mustPathString("$.name"),
		},
		meta: func(nodes []parsedNode) Meta {
			return IdentParameter(nodes[1].value.(StringMap), nodes[0].value)
		},
	},
	{
		kind: IdentKindResult,
		paths: []*yaml.Path{
			mustPathString("$.spec.results[*]"),
			mustPathString("$.name"),
		},
		meta: func(nodes []parsedNode) Meta {
			return IdentResult(nodes[1].value.(StringMap))
		},
	},
	{
		kind: IdentKindWorkspace,
		paths: []*yaml.Path{
			mustPathString("$.spec.workspaces[*]"),
			mustPathString("$.name"),
		},
		meta: func(nodes []parsedNode) Meta {
			return IdentWorkspace(nodes[1].value.(StringMap))
		},
	},
	{
		kind: IdentKindPipelineTask,
		paths: []*yaml.Path{
			mustPathString("$.spec.tasks[*]"),
			mustPathString("$.name"),
		},
		meta: func(nodes []parsedNode) Meta {
			return PipelineTask(nodes[1].value.(StringMap))
		},
	},
	{
		kind: IdentKindTask,
		paths: []*yaml.Path{
			mustPathString("$.metadata.name"),
		},
		meta: func(nodes []parsedNode) Meta {
			kind, ok := nodes[0].value.(map[string]interface{})["kind"].(string)
			if !ok || strings.ToLower(kind) != "task" {
				return nil
			}
			return IdentTask(nodes[0].value.(StringMap))
		},
	},
}

func (d *Document) getNodeRange(node ast.Node) (r protocol.Range, offsets []int) {
	r.Start = protocol.Position{
		Line:      uint32(node.GetToken().Position.Line - 1),
		Character: uint32(node.GetToken().Position.Column - 1),
	}
	startOffset := d.PositionOffset(r.Start)
	endOffset := startOffset + len(node.String())
	r.End = d.OffsetPosition(endOffset)
	offsets = []int{
		startOffset, endOffset, // "whole reference"
		startOffset, endOffset, // only the name portion
	}
	return
}

func (d *Document) parseIdentifiers() {
	d.identifiers = d.identifiers[:0]
	for _, ident := range identifiers {
		visitPath(d.ast.Body, ident.paths, func(nodes []parsedNode) {
			meta := ident.meta(nodes)
			if meta == nil {
				return
			}

			def := nodes[len(nodes)-1]
			defRange, _ := d.getNodeRange(def.node)
			id := &identifier{
				kind:       ident.kind,
				meta:       meta,
				definition: def.node,
				location: protocol.Location{
					Range: defRange,
					URI:   d.file.uri,
				},
			}
			d.identifiers = append(d.identifiers, id)
		})
	}
}

type parsedNode struct {
	node  ast.Node
	value interface{}
}

type pathFunc = func([]parsedNode)

func visitPath(node ast.Node, paths []*yaml.Path, f pathFunc, nodes ...parsedNode) {
	pn := parsedNode{
		node: node,
	}
	// go-yaml bug:
	// ```yaml
	// val: >-
	//   name
	//   # if there is no trailing space here (or another node),
	//   # the value is unmarshalled as `nam`
	// ```
	// workaround: append a whitespace before unmarshalling
	err := yaml.Unmarshal([]byte(node.String()+" "), &pn.value)
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
			visitPath(v, paths[1:], f, nodes...)
		}
	} else {
		visitPath(filtered, paths[1:], f, nodes...)
	}
}

// TODO: move to yaml package
type visitFunc = func(ast.Node)

func visitNodes(node ast.Node, depth int, f visitFunc) {
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
		visitNodes(v, depth-1, f)
	}
}
