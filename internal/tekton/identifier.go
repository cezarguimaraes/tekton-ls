package tekton

import (
	"fmt"
	"os"
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

func (d *Document) getIdent(kind identifierKind, name string) *identifier {
	for _, id := range d.identifiers {
		if id.kind != kind {
			continue
		}
		if id.meta.Name() != name {
			continue
		}
		return id
	}
	return nil
}

var identifiers = []struct {
	kind identifierKind
	meta func(StringMap) Meta

	listPath *yaml.Path
	depth    int
	namePath *yaml.Path // optional
}{
	{
		kind:     IdentKindParam,
		listPath: mustPathString("$.spec.parameters[*]"),
		depth:    1,
		meta: func(s StringMap) Meta {
			return IdentParameter(s)
		},
	},
	{
		kind:     IdentKindResult,
		listPath: mustPathString("$.spec.results[*]"),
		depth:    1,
		meta: func(s StringMap) Meta {
			return IdentResult(s)
		},
	},
	{
		kind:     IdentKindWorkspace,
		listPath: mustPathString("$.spec.workspaces[*]"),
		depth:    1,
		meta: func(s StringMap) Meta {
			return IdentWorkspace(s)
		},
	},
	{
		kind:     IdentKindPipelineTask,
		listPath: mustPathString("$.spec.tasks[*]"),
		depth:    1,
		meta: func(s StringMap) Meta {
			return PipelineTask(s)
		},
	},
	{
		kind: IdentKindTask,
		// $ path does not work, so we handle listPath == nil and don't filter
		// listPath: mustPathString("$"),
		namePath: mustPathString("$.metadata.name"),
		depth:    0,
		meta: func(s StringMap) Meta {
			kind, ok := s["kind"].(string)
			if !ok || strings.ToLower(kind) != "task" {
				return nil
			}
			return IdentTask(s)
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
		node := d.ast.Body
		var err error
		if ident.listPath != nil {
			node, err = ident.listPath.FilterNode(d.ast.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error listing ident %s: %v", ident.kind, err)
				continue
			}
		}

		identMap := make(map[string]*identifier)
		visitNodes(node, ident.depth, func(n ast.Node) {
			var v StringMap
			// go-yaml bug:
			// ```yaml
			// val: >-
			//   name
			//   # if there is no trailing space here (or another node),
			//   # the value is unmarshalled as `nam`
			// ```
			// workaround: append a whitespace before unmarshalling
			_ = yaml.Unmarshal([]byte(n.String()+" "), &v)
			meta := ident.meta(v)
			if meta == nil {
				return
			}

			namePath := ident.namePath
			if namePath == nil {
				namePath = mustPathString("$.name")
			}

			nameNode, err := namePath.FilterNode(n)
			if err != nil {
				panic("should never happen")
			}

			defRange, _ := d.getNodeRange(nameNode)
			id := &identifier{
				kind:       ident.kind,
				meta:       ident.meta(v),
				definition: nameNode,
				location: protocol.Location{
					Range: defRange,
					URI:   d.file.uri,
				},
			}
			d.identifiers = append(d.identifiers, id)

			identMap[id.meta.Name()] = id
		})
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
