package tekton

import (
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

func (f *File) findReference(pos protocol.Position) *reference {
	for _, ref := range f.references {
		// assuming ref.start.Line = ref.end.Line
		if ref.start.Line != pos.Line {
			continue
		}
		if pos.Character > ref.end.Character {
			continue
		}
		if pos.Character < ref.start.Character {
			continue
		}
		return &ref
	}
	return nil
}

type StringMap = map[string]string

func mustPathString(path string) *yaml.Path {
	p, err := yaml.PathString(path)
	if err != nil {
		panic(err)
	}
	return p
}
