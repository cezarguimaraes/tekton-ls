package tekton

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/cezarguimaraes/tekton-ls/internal/file"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type completion struct {
	context *yaml.Path
	text    string
}

type Meta interface {
	Name() string
	Documentation() string
	Completions() []completion
}

type reference struct {
	kind identifierKind
	name string

	docURI string

	offsets []int
	start   protocol.Position
	end     protocol.Position

	ident *identifier
}

type File struct {
	file.File

	workspace *Workspace

	uri string

	ast        *ast.File
	parseError error

	docs []*Document

	danglingRefs map[string]struct{}
}

type Document struct {
	file.File

	file *File

	offset int
	size   int

	ast *ast.DocumentNode

	parameters []Meta
	results    []Meta
	workspaces []Meta

	identifiers []*identifier
	references  []reference
}

var helmSanitizerRegexp = regexp.MustCompile(`{{.*?}}`)

func NewFile(f file.File) *File {
	r := &File{
		File: f,
	}

	sanitized := helmSanitizerRegexp.ReplaceAllFunc(f.Bytes(), func(src []byte) []byte {
		return []byte(strings.Repeat("x", len(src)))
	})

	r.ast, r.parseError = parser.ParseBytes(sanitized, 0)
	// document separator -- is not considered parse error
	if r.parseError != nil {
		return r
	}

	for i, doc := range r.ast.Docs {
		d := &Document{
			File: f,

			file: r,
			ast:  doc,

			offset: r.LineOffset(doc.Body.GetToken().Position.Line - 1),

			references:  []reference{},
			identifiers: []*identifier{},
		}
		if i > 0 {
			r.docs[i-1].size = d.offset - r.docs[i-1].offset
		}
		r.docs = append(r.docs, d)
	}
	if len(r.docs) > 0 {
		lst := r.docs[len(r.docs)-1]
		lst.size = len(r.Bytes()) - lst.offset
	}
	return r
}

// used only for tests
func ParseFile(f file.File) *File {
	ws := NewWorkspace()
	uri := "file://test.yaml"
	ws.UpsertFile(uri, string(f))
	ws.Lint()

	return ws.File(uri)
}

func (f *File) solveReferences() {
	if f.parseError != nil {
		return
	}
	f.danglingRefs = map[string]struct{}{}
	for _, d := range f.docs {
		d.solveReferences()
	}
}

func (f *File) solveIdentifiers() {
	if f.parseError != nil {
		return
	}
	for _, d := range f.docs {
		d.parseIdentifiers()
	}
}

func (f *File) getIdent(l identLocator) *identifier {
	ids := []*identifier{}
	for _, d := range f.docs {
		if id := d.getIdent(l); id != nil {
			ids = append(ids, id)
		}
	}
	if len(ids) > 1 {
		panic(
			fmt.Errorf(
				"file.getIdent(%v) returned more than one identifier",
				l,
			),
		)
	}
	if len(ids) == 0 {
		return nil
	}
	return ids[0]
}

func (f *File) findDoc(pos protocol.Position) *Document {
	for _, d := range f.docs {
		st := d.OffsetPosition(d.offset)
		en := d.OffsetPosition(d.offset + d.size)
		if inRange(pos, protocol.Range{Start: st, End: en}) {
			return d
		}
	}
	return nil
}

func (f *File) Hover(pos protocol.Position) *string {
	return f.findDoc(pos).hover(pos)
}

func (f *File) Rename(pos protocol.Position, newName string) (*protocol.WorkspaceEdit, error) {
	return f.findDoc(pos).rename(pos, newName)
}

func (f *File) PrepareRename(pos protocol.Position) *protocol.Location {
	return f.findDoc(pos).prepareRename(pos)
}

func (f *File) Definition(pos protocol.Position) *protocol.Location {
	return f.findDoc(pos).definition(pos)
}

func (f *File) FindReferences(pos protocol.Position) []protocol.Location {
	return f.findDoc(pos).findReferences(pos)
}

func (f *File) Completions(pos protocol.Position) []fmt.Stringer {
	res := []fmt.Stringer{}
	if f.parseError != nil {
		return res
	}
	return f.findDoc(pos).completions(pos)
}

func (f *File) Diagnostics() []protocol.Diagnostic {
	if f.parseError != nil {
		if d := syntaxErrorDiagnostic(f.parseError); d != nil {
			return []protocol.Diagnostic{*d}
		}
	}

	dgs := make(chan *protocol.Diagnostic, len(f.docs))
	var wg sync.WaitGroup
	for _, d := range f.docs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.diagnostics(dgs)
		}()
	}

	go func() {
		wg.Wait()
		close(dgs)
	}()

	rs := make([]protocol.Diagnostic, 0, len(f.docs))
	for dg := range dgs {
		rs = append(rs, *dg)
	}

	return rs
}

type StringMap = map[string]interface{}

func mustPathString(path string) *yaml.Path {
	p, err := yaml.PathString(path)
	if err != nil {
		panic(err)
	}
	return p
}
