package tekton

import (
	"fmt"
	"os"
	"regexp"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
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
	references [][]int
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

func (f *File) parseIdentifiers() []*identifier {
	ids := []*identifier{}
	for _, ident := range identifiers {
		node, err := ident.listPath.FilterFile(f.ast)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error listing ident %s: %v", ident.kind, err)
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
			).FilterFile(f.ast)

			id := &identifier{
				kind:       ident.kind,
				meta:       ident.meta(m),
				definition: def,
			}
			ids = append(ids, id)

			identMap[id.meta.Name()] = id
		}

		refs := ident.referenceRegex.FindAllSubmatchIndex(f.Bytes(), 1000)
		for _, match := range refs {
			name := string(f.Bytes())[match[2]:match[3]]
			id, _ := identMap[name]
			if id != nil {
				id.references = append(id.references, match)
			}
			f.references = append(f.references, reference{
				kind:    ident.kind,
				name:    name,
				ident:   id,
				start:   f.OffsetPosition(match[0]),
				end:     f.OffsetPosition(match[1]),
				offsets: match,
			})
		}
	}
	return ids
}
