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

var identifiers = []struct {
	kind           identifierKind
	listPath       *yaml.Path
	pathFormat     string
	referenceRegex *regexp.Regexp
	meta           func(StringMap) Meta
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
	},
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

			start := protocol.Position{
				Line:      uint32(def.GetToken().Position.Line - 1),
				Character: uint32(def.GetToken().Position.Column - 1),
			}
			id := &identifier{
				kind:       ident.kind,
				meta:       ident.meta(m),
				definition: def,
				prange: protocol.Range{
					Start: start,
					// end is calculated in this roundabount way due to
					// inconsistency of Token.Position.Offset for
					// string token
					End: d.OffsetPosition(
						d.PositionOffset(start) + len(def.String()),
					),
				},
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
	}
	d.identifiers = ids
}
