package tekton

import (
	"github.com/cezarguimaraes/tekton-lsp/internal/file"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

type Meta interface {
	Name() string
	Documentation() string
	Completions() []string
}

type reference struct {
	kind identifierKind
	name string

	offsets []int

	ident *identifier
}

type File struct {
	file.File

	ast        *ast.File
	parseError error

	parameters []Meta
	results    []Meta
	workspaces []Meta

	identifiers []*identifier
	references  []reference
}

func ParseFile(f file.File) *File {
	r := &File{
		File: f,
	}

	r.ast, r.parseError = parser.ParseBytes(f.Bytes(), 0)
	if r.parseError != nil {
		return r
	}

	r.identifiers = r.parseIdentifiers()

	return r
}

type StringMap = map[string]string

func mustPathString(path string) *yaml.Path {
	p, err := yaml.PathString(path)
	if err != nil {
		panic(err)
	}
	return p
}
