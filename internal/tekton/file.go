package tekton

import (
	"fmt"

	"github.com/cezarguimaraes/tekton-lsp/internal/file"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	protocol "github.com/tliron/glsp/protocol_3_16"
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
	start   protocol.Position
	end     protocol.Position

	ident *identifier
}

type File struct {
	file.File

	ast        *ast.File
	parseError error

	docs []*Document
}

type Document struct {
	file.File

	ast *ast.DocumentNode

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

	for _, doc := range r.ast.Docs {
		d := &Document{
			File: f,
			ast:  doc,
		}
		d.parseIdentifiers()

		r.docs = append(r.docs, d)
	}

	return r
}

func (f *File) findDoc(pos protocol.Position) *Document {
	for _, d := range f.docs {
		if d.ast.End == nil || pos.Line < uint32(d.ast.End.Position.Line) {
			return d
		}
	}
	return nil
}

func (f *File) Hover(pos protocol.Position) *string {
	return f.findDoc(pos).hover(pos)
}

func (f *File) Definition(pos protocol.Position) *protocol.Range {
	return f.findDoc(pos).definition(pos)
}

func (f *File) FindReferences(pos protocol.Position) []protocol.Range {
	return f.findDoc(pos).findReferences(pos)
}

func (f *File) Completions(pos protocol.Position) []fmt.Stringer {
	res := []fmt.Stringer{}
	if f.parseError != nil {
		return res
	}
	return f.findDoc(pos).completions()
}

func (f *File) Diagnostics() []protocol.Diagnostic {
	rs := make([]protocol.Diagnostic, 0)
	if f.parseError != nil {
		if d := syntaxErrorDiagnostic(f.parseError); d != nil {
			rs = append(rs, *d)
			return rs
		}
	}
	for _, d := range f.docs {
		rs = append(rs, d.diagnostics()...)
	}
	return rs
}

type StringMap = map[string]string

func mustPathString(path string) *yaml.Path {
	p, err := yaml.PathString(path)
	if err != nil {
		panic(err)
	}
	return p
}
