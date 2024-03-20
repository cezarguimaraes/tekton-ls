package tekton

import (
	"regexp"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type referenceResolver interface {
	find(*Document)
}

type regexpRef struct {
	kind  identifierKind
	regex *regexp.Regexp
}

var _ referenceResolver = &regexpRef{}

func (r *regexpRef) find(d *Document) {
	// this can be reused between documents
	refs := r.regex.FindAllSubmatchIndex(d.Bytes(), 1000)
	for _, match := range refs {
		name := string(d.Bytes())[match[2]:match[3]]
		id := d.getIdent(r.kind, name)

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
			kind:    r.kind,
			name:    name,
			ident:   id,
			start:   start,
			end:     end,
			offsets: match,
		})
	}
}

type pathRef struct {
	path    *yaml.Path
	depth   int
	handler func(*Document, interface{}, ast.Node) []reference
}

var _ referenceResolver = &pathRef{}

func (r *pathRef) find(d *Document) {
	node, err := r.path.FilterNode(d.ast.Body)
	if err != nil {
		panic(err)
	}

	visitNodes(node, r.depth, func(n ast.Node) {
		var v interface{}
		_ = yaml.Unmarshal([]byte(n.String()), &v)

		refs := r.handler(d, v, n)
		for _, ref := range refs {
			if id := ref.ident; id != nil {
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
			d.references = append(d.references, ref)
		}
	})
}

var references = []referenceResolver{
	&regexpRef{
		kind:  IdentKindParam,
		regex: regexp.MustCompile(`\$\(params\.(.*?)\)`),
	},
	&regexpRef{
		kind:  IdentKindResult,
		regex: regexp.MustCompile(`\$\(results\.(.*?)\.(.*?)\)`),
	},
	&regexpRef{
		kind:  IdentKindWorkspace,
		regex: regexp.MustCompile(`\$\(workspaces\.(.*?)\.(.*?)\)`),
	},
	&regexpRef{
		kind:  IdentKindPipelineTask,
		regex: regexp.MustCompile(`\$\(tasks\.(.*?)\.(.*?)\.(.*?)\)`),
	},
	&pathRef{
		path:  mustPathString("$.spec.tasks[*].workspaces[*]"),
		depth: 2,
		handler: func(d *Document, v interface{}, node ast.Node) []reference {
			vm := v.(map[string]interface{})
			ws, ok := vm["workspace"]
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
					kind:    IdentKindWorkspace,
					name:    wsName,
					ident:   d.getIdent(IdentKindWorkspace, wsName),
					start:   prange.Start,
					end:     prange.End,
					offsets: offsets,
				},
			}
		},
	},
	&pathRef{
		path:  mustPathString("$.spec.tasks[*].runAfter[*]"),
		depth: 2,
		handler: func(d *Document, v interface{}, node ast.Node) []reference {
			s, ok := v.(string)
			if !ok {
				return nil
			}
			prange, offsets := d.getNodeRange(node)
			return []reference{
				{
					kind:    IdentKindPipelineTask,
					name:    s,
					ident:   d.getIdent(IdentKindPipelineTask, s),
					start:   prange.Start,
					end:     prange.End,
					offsets: offsets,
				},
			}
		},
	},
	&pathRef{
		path:  mustPathString("$.spec.tasks[*].taskRef.name"),
		depth: 1,
		handler: func(d *Document, v interface{}, node ast.Node) []reference {
			s, ok := v.(string)
			if !ok {
				return nil
			}
			prange, offsets := d.getNodeRange(node)
			return []reference{
				{
					kind:    IdentKindTask,
					name:    s,
					ident:   d.file.getIdent(IdentKindTask, s),
					start:   prange.Start,
					end:     prange.End,
					offsets: offsets,
				},
			}
		},
	},
	&pathRef{
		path:  mustPathString("$.spec.tasks[*].parameters[*].name"),
		depth: 2,
		handler: func(d *Document, v interface{}, node ast.Node) []reference {
			s, ok := v.(string)
			if !ok {
				return nil
			}
			prange, offsets := d.getNodeRange(node)
			return []reference{
				{
					kind: IdentKindParam,
					name: s,
					// TODO: get param from the correct task :)
					ident:   d.file.getIdent(IdentKindParam, s),
					start:   prange.Start,
					end:     prange.End,
					offsets: offsets,
				},
			}
		},
	},
}

func wholeReferences(id *identifier) []protocol.Range {
	if id == nil {
		return nil
	}
	var refs []protocol.Range
	for _, ref := range id.references {
		refs = append(refs, ref[0])
	}
	return refs
}

func (d *Document) solveReferences() {
	for _, ref := range references {
		ref.find(d)
	}
}

func (d *Document) findIdentifier(pos protocol.Position) *identifier {
	for _, id := range d.identifiers {
		if inRange(pos, id.prange) {
			return id
		}
	}
	return nil
}

func (d *Document) findReferences(pos protocol.Position) []protocol.Range {
	return wholeReferences(d.findIdentifier(pos))
}

func cmpPos(a, b protocol.Position) bool {
	if a.Line < b.Line {
		return true
	}
	if a.Line == b.Line && a.Character < b.Character {
		return true
	}
	return false
}

func inRange(pos protocol.Position, r protocol.Range) bool {
	// r.start <= pos && pos < r.end
	return !cmpPos(pos, r.Start) && cmpPos(pos, r.End)
}
