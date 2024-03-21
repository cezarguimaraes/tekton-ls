package tekton

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	var wg sync.WaitGroup
	for _, f := range w.files {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.solveIdentifiers()
		}()
	}
	wg.Wait()
	for _, f := range w.files {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.solveReferences()
		}()
	}
	wg.Wait()
}

func (w *Workspace) AddFolder(uri string) {
	base := strings.TrimPrefix(uri, "file://")
	c := make(chan *File)
	go func() {
		var wg sync.WaitGroup
		filepath.WalkDir(
			base,
			func(path string, de fs.DirEntry, err error) error {
				if de.IsDir() {
					return nil
				}
				if err != nil {
					return nil
				}
				ext := filepath.Ext(path)
				if ext != ".yaml" && ext != ".yml" {
					return nil
				}

				wg.Add(1)
				go func() {
					defer wg.Done()
					d, err := os.ReadFile(path)
					if err != nil {
						// TODO: report errors
						return
					}
					uri := "file://" + path
					f := NewFile(file.File(string(d)))
					f.uri = uri
					f.workspace = w
					c <- f
				}()
				return nil
			},
		)
		wg.Wait()
		close(c)
	}()

	for f := range c {
		w.files[f.uri] = f
	}
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

func (w *Workspace) Diagnostics(cb func(protocol.PublishDiagnosticsParams)) {
	var wg sync.WaitGroup
	for _, f := range w.files {
		// TODO: keep track of and include file version
		go func() {
			wg.Add(1)
			defer wg.Done()
			cb(protocol.PublishDiagnosticsParams{
				URI:         f.uri,
				Diagnostics: f.Diagnostics(),
			})
		}()
	}
	wg.Wait()
}
