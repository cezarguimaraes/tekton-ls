package tekton

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cezarguimaraes/tekton-ls/internal/file"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type Workspace struct {
	files map[string]*File
}

func NewWorkspace() *Workspace {
	return &Workspace{
		files: make(map[string]*File),
	}
}

func (w *Workspace) File(uri string) *File {
	return w.files[uri]
}

func (w *Workspace) UpsertFile(uri string, text string) {
	f := NewFile(file.File(text))
	f.workspace = w
	f.uri = uri
	w.files[uri] = f
}

func (w *Workspace) Lint() {
	for _, f := range w.files {
		f.solveIdentifiers()
	}
	for _, f := range w.files {
		f.solveReferences()
	}
}

func (w *Workspace) AddFolder(uri string) {
	base := strings.TrimPrefix(uri, "file://")
	// TODO: change to walkdir
	filepath.Walk(
		base,
		func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if err != nil {
				return nil
			}
			ext := filepath.Ext(path)
			if ext != ".yaml" && ext != ".yml" {
				return nil
			}
			d, err := os.ReadFile(path)
			if err != nil {
				// TODO: report errors
				return nil
			}
			uri := "file://" + path
			w.UpsertFile(uri, string(d))
			return nil
		},
	)
}

// TODO: add diagnostic for when there are multiple idents
func (w *Workspace) getIdent(l identLocator) *identifier {
	for _, f := range w.files {
		id := f.getIdent(l)
		if id != nil {
			return id
		}
	}
	return nil
}

func (w *Workspace) FindReferences(docUri string, pos protocol.Position) []protocol.Location {
	return w.File(docUri).FindReferences(pos)
}

func (w *Workspace) Rename(docUri string, pos protocol.Position, newName string) (*protocol.WorkspaceEdit, error) {
	return w.File(docUri).Rename(pos, newName)
}

func (w *Workspace) Diagnostics() []protocol.PublishDiagnosticsParams {
	var dgs []protocol.PublishDiagnosticsParams
	for _, f := range w.files {
		dgs = append(dgs, protocol.PublishDiagnosticsParams{
			URI:         f.uri,
			Diagnostics: f.Diagnostics(),
		})
	}
	return dgs
}
