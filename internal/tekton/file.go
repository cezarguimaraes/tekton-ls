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

type StringMap = map[string]interface{}

// Meta is the root interface for any given Tekton object understood by this
// language server.
type Meta interface {
	// Name returns the name of the given Tekton object.
	Name() string

	// Documentation returns the Markdown documentation of the given Tekton
	// object.
	Documentation() string

	// Completions returns a list of possible completion suggestions
	// associated with the Tekton object.
	Completions() []completion
}

// File provides operation on a YAML file containing any number of
// Tekton Documents - separated by `---` as defined by the YAML spec.
type File struct {
	// TextDocument of this file.
	file.TextDocument

	// Workspace containing this file.
	workspace *Workspace

	// uri is the TextDocument URI of this file.
	uri string

	// ast is the abstract syntax tree of this YAML file.
	ast        *ast.File
	parseError error

	// docs is the list of tekton.Document contained in this file.
	docs []*Document

	// danglingRefs is a set containing the name of any reference to which
	// no identifier could be mapped. This set is used to identify which
	// files have to be re-linted whenever a file in the workspace is changed.
	danglingRefs map[string]struct{}
}

// Document provides operation on a single YAML document containing a Tekton
// resource.
type Document struct {
	// TextDocument is the LSP TextDocument which contains this document.
	file.TextDocument

	// file containing this document.
	file *File

	// offset is the starting position of the Document in its containing File.
	offset int
	// size of the document in its containing File.
	size int

	// ast is the abstract syntax tree of this YAML document.
	ast *ast.DocumentNode

	// identifiers is the list of identifiers (i.e definitions) in this file.
	identifiers []*identifier

	// references is the list of possible references to identifiers in this file.
	references []reference
}

// helmSanitizerRegexp is the regular expression used to identify Helm template
// language directives. These directives are filtered out of the file contents
// to prevent YAML syntax errors.
var helmSanitizerRegexp = regexp.MustCompile(`{{.*?}}`)

// NewFile sanitizes a TextDocument, parses its YAML contents into an AST,
// and calculates the offset and size of every YAML document in the file.
func NewFile(f file.TextDocument) *File {
	r := &File{
		TextDocument: f,
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
			TextDocument: f,

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

// deprecated: used only for texts
func parseFile(f file.TextDocument) *File {
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

// getIdent accepts an identLocator and returns a pointer to an identifier
// in this File matched by the locator, or nil if none matches.
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

// findDoc returns the Document containing the given position.
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

// Hover returns a pointer to a string containing the Documentation of the
// Tekton object in the given position, or nil if no reference is found.
func (f *File) Hover(pos protocol.Position) *string {
	return f.findDoc(pos).hover(pos)
}

// Rename returns the Workspace changes required to rename the identifier
// in the given position to the newName argument.
func (f *File) Rename(pos protocol.Position, newName string) (*protocol.WorkspaceEdit, error) {
	return f.findDoc(pos).rename(pos, newName)
}

// PrepareRename returns the Location (File, Range) that should be edited
// if a Rename request is made at the given position. It should return
// nil if it's not a valid position for a rename request, that is, if no
// identifier is found at the given position.
func (f *File) PrepareRename(pos protocol.Position) *protocol.Location {
	return f.findDoc(pos).prepareRename(pos)
}

// Definition returns the Location where the identifier in the given position
// is defined, or nil if no identifier is found.
func (f *File) Definition(pos protocol.Position) *protocol.Location {
	return f.findDoc(pos).definition(pos)
}

// FindReferences returns a list of all Locations which refer to the identifier
// in the given position, or nil if no identifier is found.
func (f *File) FindReferences(pos protocol.Position) []protocol.Location {
	return f.findDoc(pos).findReferences(pos)
}

// Completions returns a list of completion suggestions for the given
// position.
func (f *File) Completions(pos protocol.Position) []fmt.Stringer {
	res := []fmt.Stringer{}
	if f.parseError != nil {
		return res
	}
	return f.findDoc(pos).completions(pos)
}

// Diagnostics returns a list of Diagnostics issues found in this File.
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

func mustPathString(path string) *yaml.Path {
	p, err := yaml.PathString(path)
	if err != nil {
		panic(err)
	}
	return p
}
