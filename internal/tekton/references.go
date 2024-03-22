package tekton

import (
	"regexp"

	yaml_helper "github.com/cezarguimaraes/tekton-ls/internal/yaml"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// reference holds information about a text fragment which purposes to refer
// to an identifier.
type reference struct {
	// kind is the identifier kind of this reference.
	kind identifierKind

	// name is the identifier name referred to by this reference.
	name string

	// docURI is the TextDocument URI in which this reference is found.
	docURI string

	// offsets is an array of [start, end) text positions identified in this
	// reference. For example, for a reference of the kind $(params.name),
	// text[offsets[0]:offsets[1]] contains `$(params.name)` and
	// text[offset[2]:offsets[3]] contains `name`.
	offsets []int

	// start contains the inclusive start position of this reference in
	// the document.
	start protocol.Position
	// end contains the exclusive end position of this reference in
	// the document.
	end protocol.Position

	// ident is the identifier found to be referred by this reference. It can
	// be nil if no identifier of the given name is found.
	ident *identifier
}

// referenceResolver is an interface able to locate some kind of references
// in a given Tekton Document. Additionally from identifying references,
// it should query the appropriate Tekton abstraction (among Workspace, File,
// and Document) for the identifier referred by the references it finds.
// It additionally must update the document.danglingRefs set with the name
// of any identifier it isn't able to find.
type referenceResolver interface {
	find(*Document)
}

// regexpRef implements referenceResolver given a regular expression.
type regexpRef struct {
	kind  identifierKind
	regex *regexp.Regexp
}

var _ referenceResolver = &regexpRef{}

// find implements referenceResolver by finding all matches of a regular
// expression in a given document.
func (r *regexpRef) find(d *Document) {
	// this can be reused between documents
	refs := r.regex.FindAllSubmatchIndex(d.Bytes(), 1000)
	for _, match := range refs {
		name := string(d.Bytes())[match[2]:match[3]]
		id := d.getIdent(&kindNameLocator{r.kind, name})

		if match[0] < d.offset || match[1] > d.offset+d.size {
			continue
		}

		start := d.OffsetPosition(match[0])
		end := d.OffsetPosition(match[1])
		if id != nil {
			id.references = append(id.references, []protocol.Location{
				{
					URI: d.file.uri,
					Range: protocol.Range{
						Start: start,
						End:   end,
					},
				},
				{
					URI: d.file.uri,
					Range: protocol.Range{
						Start: d.OffsetPosition(match[2]),
						End:   d.OffsetPosition(match[3]),
					},
				},
			})
		} else {
			d.file.danglingRefs[name] = struct{}{}
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

// deprecated: move to pathRef2
type pathRef struct {
	path    *yaml.Path
	depth   int
	handler func(*Document, interface{}, ast.Node) []reference
}

var _ referenceResolver = &pathRef{}

func (r *pathRef) find(d *Document) {
	node, err := r.path.FilterNode(d.ast.Body)
	if err != nil {
		return
	}

	yaml_helper.VisitNodes(node, r.depth, func(n ast.Node) {
		var v interface{}
		_ = yaml.Unmarshal([]byte(n.String()), &v)

		refs := r.handler(d, v, n)
		for _, ref := range refs {
			ref.docURI = d.file.uri
			if id := ref.ident; id != nil {
				id.references = append(id.references, []protocol.Location{
					{
						URI: d.file.uri,
						Range: protocol.Range{
							Start: ref.start,
							End:   ref.end,
						},
					},
					{
						URI: d.file.uri,
						Range: protocol.Range{
							Start: ref.start,
							End:   ref.end,
						},
					},
				})
			} else {
				d.file.danglingRefs[ref.name] = struct{}{}
			}

			d.references = append(d.references, ref)
		}
	})
}

// pathRef2 implements referenceResolver given a recursive YAML path list
// and a function handler which turns the list of parsedNode found by
// yaml.VisitPath into a list of references. Check yaml.VisitPath for more
// information.
type pathRef2 struct {
	paths   []*yaml.Path
	handler func(*Document, []yaml_helper.ParsedNode) []reference
}

var _ referenceResolver = &pathRef2{}

func (r *pathRef2) find(d *Document) {
	yaml_helper.VisitPath(d.ast.Body, r.paths, func(pn []yaml_helper.ParsedNode) {
		refs := r.handler(d, pn)
		for _, ref := range refs {
			ref.docURI = d.file.uri
			if id := ref.ident; id != nil {
				loc := protocol.Location{
					URI: d.file.uri,
					Range: protocol.Range{
						Start: ref.start,
						End:   ref.end,
					},
				}

				id.references = append(
					id.references,
					[]protocol.Location{loc, loc},
				)
			} else {
				d.file.danglingRefs[ref.name] = struct{}{}
			}
			d.references = append(d.references, ref)
		}
	})
}

// references is the list of referenceResolver used to find all references
// in a given Tekton Document.
var references = []referenceResolver{
	&regexpRef{
		kind:  IdentKindParam,
		regex: regexp.MustCompile(`\$\(params\.(.*?)(\[\*\])?\)`),
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
				return nil
			}

			prange, offsets := d.getNodeRange(nameNode)
			return []reference{
				{
					kind:    IdentKindWorkspace,
					name:    wsName,
					ident:   d.getIdent(&kindNameLocator{IdentKindWorkspace, wsName}),
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
					ident:   d.getIdent(&kindNameLocator{IdentKindPipelineTask, s}),
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
					ident:   d.file.workspace.getIdent(&kindNameLocator{IdentKindTask, s}),
					start:   prange.Start,
					end:     prange.End,
					offsets: offsets,
				},
			}
		},
	},
	&pathRef2{
		paths: []*yaml.Path{
			mustPathString("$.spec.tasks[*]"),
			mustPathString("$.params[*]"),
			mustPathString("$.name"),
			// mustPathString("$.spec.tasks[*].params[*].name"),
		},
		handler: func(d *Document, nodes []yaml_helper.ParsedNode) []reference {
			s, ok := nodes[3].Value.(string)
			if !ok {
				return nil
			}

			parent := nodes[1].Value.(map[string]interface{})
			tr := parent["taskRef"]
			var trm map[string]interface{}
			if tr != nil {
				trm, _ = tr.(map[string]interface{})
			}
			var taskName string
			if trm != nil {
				ni := trm["name"]
				if ni != nil {
					taskName, _ = ni.(string)
				}
			}

			prange, offsets := d.getNodeRange(nodes[3].Node)
			return []reference{
				{
					kind: IdentKindParam,
					name: s,
					// TODO: get param from the correct task :)
					ident: d.file.workspace.getIdent(&taskParamLocator{
						name:     s,
						taskName: taskName,
					}),
					start:   prange.Start,
					end:     prange.End,
					offsets: offsets,
				},
			}
		},
	},
}

// wholeReferences returns the largest Range which identifies a reference for
// all references found for a given identifier. Check reference.offsets for
// more information.
func wholeReferences(id *identifier) []protocol.Location {
	if id == nil {
		return nil
	}
	var refs []protocol.Location
	for _, ref := range id.references {
		refs = append(refs, ref[0])
	}
	return refs
}

func (d *Document) solveReferences() {
	d.references = d.references[:0]
	for _, ref := range references {
		ref.find(d)
	}
}

func (d *Document) findIdentifier(pos protocol.Position) *identifier {
	for _, id := range d.identifiers {
		if inRange(pos, id.location.Range) {
			return id
		}
	}
	return nil
}

func (d *Document) findReferences(pos protocol.Position) []protocol.Location {
	return wholeReferences(d.findIdentifier(pos))
}

// cmpPos provides a strict partial order for TextDocument positions.
func cmpPos(a, b protocol.Position) bool {
	if a.Line < b.Line {
		return true
	}
	if a.Line == b.Line && a.Character < b.Character {
		return true
	}
	return false
}

// inRange returns true if and only if the given Position is contained by
// the given Range.
func inRange(pos protocol.Position, r protocol.Range) bool {
	// r.start <= pos && pos < r.end
	return !cmpPos(pos, r.Start) && cmpPos(pos, r.End)
}
