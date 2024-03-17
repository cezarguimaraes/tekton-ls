package tekton

import (
	"github.com/cezarguimaraes/tekton-lsp/internal/file"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

type StringMap = map[string]string

type Meta interface {
	Name() string
	Documentation() string
}

func mustPathString(path string) *yaml.Path {
	p, err := yaml.PathString(path)
	if err != nil {
		panic(err)
	}
	return p
}

type File struct {
	file.File

	ast        *ast.File
	parseError error

	parameters []Meta
	results    []Meta
	workspaces []Meta
}

func ParseFile(f file.File) *File {
	r := &File{
		File: f,
	}

	r.ast, r.parseError = parser.ParseBytes(f.Bytes(), 0)
	if r.parseError != nil {
		return r
	}

	// TODO: debug log
	r.parseParameters()
	r.parseResults()
	r.parseWorkspaces()

	return r
}

func (f *File) readPath(p *yaml.Path, v interface{}) error {
	node, err := p.FilterFile(f.ast)
	if err != nil {
		return err
	}
	return yaml.Unmarshal([]byte(node.String()), v)
}
