package tekton

import (
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"

	yaml_helper "github.com/cezarguimaraes/tekton-ls/internal/yaml"
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

// identifier holds information about any identifiers - Tekton objects that
// can be referred to.
type identifier struct {
	kind identifierKind
	meta Meta

	definition ast.Node

	location   protocol.Location
	references [][]protocol.Location
}

// identLocator defines a strategy to locate identifiers.
type identLocator interface {
	matches(*identifier) bool
}

// kindNameLocator locates an identifier given a kind and name.
type kindNameLocator struct {
	kind identifierKind
	name string
}

func (l *kindNameLocator) matches(id *identifier) bool {
	return id.kind == l.kind && id.meta.Name() == l.name
}

// taskParamLocator locates an identifier given a paramater name and the task
// which defines it.
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

// identifiers is the list of rules used to find an identifier in a given
// YAML document.
var identifiers = []struct {
	// kind is the identifier kind which this rule finds.
	kind identifierKind

	// paths is a list of recursive YAML paths required to properly construct
	// an identifier.
	paths []*yaml.Path

	// meta is the function handler which should construct a Meta Tekton object
	// given the list of nodes matched by `paths`. Check `yaml.VisitPath` for
	// more information.
	meta func([]yaml_helper.ParsedNode) Meta
}{
	{
		kind: IdentKindParam,
		paths: []*yaml.Path{
			mustPathString("$.spec.params[*]"),
			mustPathString("$.name"),
		},
		meta: func(nodes []yaml_helper.ParsedNode) Meta {
			return IdentParameter(nodes[1].Value.(StringMap), nodes[0].Value)
		},
	},
	{
		kind: IdentKindResult,
		paths: []*yaml.Path{
			mustPathString("$.spec.results[*]"),
			mustPathString("$.name"),
		},
		meta: func(nodes []yaml_helper.ParsedNode) Meta {
			return IdentResult(nodes[1].Value.(StringMap))
		},
	},
	{
		kind: IdentKindWorkspace,
		paths: []*yaml.Path{
			mustPathString("$.spec.workspaces[*]"),
			mustPathString("$.name"),
		},
		meta: func(nodes []yaml_helper.ParsedNode) Meta {
			return IdentWorkspace(nodes[1].Value.(StringMap))
		},
	},
	{
		kind: IdentKindPipelineTask,
		paths: []*yaml.Path{
			mustPathString("$.spec.tasks[*]"),
			mustPathString("$.name"),
		},
		meta: func(nodes []yaml_helper.ParsedNode) Meta {
			return PipelineTask(nodes[1].Value.(StringMap))
		},
	},
	{
		kind: IdentKindTask,
		paths: []*yaml.Path{
			mustPathString("$.metadata.name"),
		},
		meta: func(nodes []yaml_helper.ParsedNode) Meta {
			kind, ok := nodes[0].Value.(map[string]interface{})["kind"].(string)
			if !ok || strings.ToLower(kind) != "task" {
				return nil
			}
			return IdentTask(nodes[0].Value.(StringMap))
		},
	},
}

// getNodeRange returns the text document Range (start, end) and offsets (as
// defined by reference.offsets) of the given AST node.
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

// parseIdentifiers locates all Tekton object identifiers (or definitions)
// in the Document. The identifier `references` property is not populated
// yet.
func (d *Document) parseIdentifiers() {
	d.identifiers = d.identifiers[:0]
	for _, ident := range identifiers {
		yaml_helper.VisitPath(d.ast.Body, ident.paths, func(nodes []yaml_helper.ParsedNode) {
			meta := ident.meta(nodes)
			if meta == nil {
				return
			}

			def := nodes[len(nodes)-1]
			defRange, _ := d.getNodeRange(def.Node)
			id := &identifier{
				kind:       ident.kind,
				meta:       meta,
				definition: def.Node,
				location: protocol.Location{
					Range: defRange,
					URI:   d.file.uri,
				},
			}
			d.identifiers = append(d.identifiers, id)
		})
	}
}
