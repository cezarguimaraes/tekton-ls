package tekton

import (
	"fmt"
	"os"
	"regexp"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type identifierKind int

const (
	IdentParam identifierKind = iota
	IdentResult
	IdentWorkspace
)

func (k identifierKind) String() string {
	switch k {
	case IdentParam:
		return "parameter"
	case IdentResult:
		return "result"
	case IdentWorkspace:
		return "workspace"
	}
	return ""
}

type identifier struct {
	kind       identifierKind
	meta       Meta
	definition ast.Node
	prange     protocol.Range
	references [][]protocol.Range
}

type identRefFunc = func(*Document, map[string]interface{}, ast.Node) []reference

type identRefs struct {
	path    *yaml.Path
	handler identRefFunc
}

var identifiers = []struct {
	kind identifierKind
	meta func(StringMap) Meta

	listPath       *yaml.Path
	pathFormat     string
	referenceRegex *regexp.Regexp

	// paths where this kind of ident might be referred
	references []identRefs
}{
	{
		kind:           IdentParam,
		listPath:       mustPathString("$.spec.parameters[*]"),
		pathFormat:     "$.spec.parameters[%d].name",
		referenceRegex: regexp.MustCompile(`\$\(params\.(.*?)\)`),
		meta: func(s StringMap) Meta {
			return Parameter(s)
		},
	},
	{
		kind:           IdentResult,
		listPath:       mustPathString("$.spec.results[*]"),
		pathFormat:     "$.spec.results[%d].name",
		referenceRegex: regexp.MustCompile(`\$\(results\.(.*?)\.(.*?)\)`),
		meta: func(s StringMap) Meta {
			return Result(s)
		},
	},
	{
		kind:           IdentWorkspace,
		listPath:       mustPathString("$.spec.workspaces[*]"),
		pathFormat:     "$.spec.workspaces[%d].name",
		referenceRegex: regexp.MustCompile(`\$\(workspaces\.(.*?)\.(.*?)\)`),
		meta: func(s StringMap) Meta {
			return Workspace(s)
		},
		references: []identRefs{
			{
				path: mustPathString("$.spec.tasks[*].workspaces[*]"),
				handler: func(d *Document, v map[string]interface{}, node ast.Node) []reference {
					ws, ok := v["workspace"]
					if !ok {
						return nil
					}
					wsName, ok := ws.(string)
					if !ok {
						return nil
					}
					// TODO: move elsewhere
					namePath := mustPathString("$.workspace")

					nameNode, err := namePath.FilterNode(node)
					if err != nil {
						panic("should never happen")
					}

					prange, offsets := d.getNodeRange(nameNode)
					return []reference{
						{
							kind:    IdentWorkspace,
							name:    wsName,
							ident:   nil,
							start:   prange.Start,
							end:     prange.End,
							offsets: offsets,
						},
					}
				},
			},
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
	ids := []*identifier{}
	for _, ident := range identifiers {
		node, err := ident.listPath.FilterNode(d.ast.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error listing ident %s: %v", ident.kind, err)
			continue
		}

		if node == nil {
			continue
		}

		ms := []StringMap{}
		if err = yaml.Unmarshal([]byte(node.String()+" "), &ms); err != nil {
			// apparent go-yaml bug:
			// ```yaml
			// val: >-
			//   name
			//   # if there is no trailing space here (or another node),
			//   # the value is unmarshalled as `nam`
			// ```
			// workaround: append a whitespace before unmarshalling

			fmt.Fprintf(os.Stderr, "error unmarshalling nodes of kind %d: %v", ident.kind, err)
			continue
		}

		identMap := make(map[string]*identifier)

		for idx, m := range ms {
			// TODO: handle nameless idents

			def, _ := mustPathString(
				fmt.Sprintf(ident.pathFormat, idx),
			).FilterNode(d.ast.Body)

			defRange, _ := d.getNodeRange(def)
			id := &identifier{
				kind:       ident.kind,
				meta:       ident.meta(m),
				definition: def,
				prange:     defRange,
			}
			ids = append(ids, id)

			identMap[id.meta.Name()] = id
		}

		// this can be reused between documents
		refs := ident.referenceRegex.FindAllSubmatchIndex(d.Bytes(), 1000)
		for _, match := range refs {
			name := string(d.Bytes())[match[2]:match[3]]
			id, _ := identMap[name]

			if match[0] < d.offset || match[1] > d.offset+d.size {
				continue
			}

			start := d.OffsetPosition(match[0])
			end := d.OffsetPosition(match[1])
			if id != nil {
				id.references = append(id.references, []protocol.Range{
					{
						Start: start,
						End:   end,
					},
					{
						Start: d.OffsetPosition(match[2]),
						End:   d.OffsetPosition(match[3]),
					},
				})
			}
			d.references = append(d.references, reference{
				kind:    ident.kind,
				name:    name,
				ident:   id,
				start:   start,
				end:     end,
				offsets: match,
			})
		}

		for _, iref := range ident.references {
			// path, handler
			node, err := iref.path.FilterNode(d.ast.Body)
			if err != nil {
				panic(err)
			}
			if node == nil {
				continue
			}
			// var vs []map[string]interface{}
			switch n := node.(type) {
			case *ast.SequenceNode:
				fmt.Println(n.Type())
				for taskId, v := range n.Values {
					if v == nil {
						// filter can leave null values
						continue
					}
					for idx, ws := range v.(*ast.SequenceNode).Values {
						_ = taskId
						_ = idx
						if ws == nil {
							panic("shouldnt happen")
						}
						var v map[string]interface{}
						_ = yaml.Unmarshal([]byte(ws.String()), &v)
						refs := iref.handler(d, v, ws)
						for _, ref := range refs {
							id, _ := identMap[ref.name]
							if id != nil {
								id.references = append(id.references, []protocol.Range{
									{
										Start: ref.start,
										End:   ref.end,
									},
									{
										Start: ref.start,
										End:   ref.end,
									},
								})
							}
							ref.ident = id
							d.references = append(d.references, ref)
						}
					}
				}
			default:
				panic("unhandled node type: " + n.Type().String())

			}

		}
	}
	d.identifiers = ids
}
