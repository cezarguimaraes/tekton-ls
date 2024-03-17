package tekton

import (
	"fmt"
	"strings"

	"github.com/cezarguimaraes/tekton-lsp/internal/file"
)

var workspacesPath = mustPathString("$.spec.workspaces[*]")

type Workspace StringMap

func (p Workspace) Name() string {
	return StringMap(p)["name"]
}

func (p Workspace) Description() string {
	return StringMap(p)["description"]
}

func (p Workspace) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\n%s\n```",
		p.Name(),
		p.Description(),
	)
}

func Workspaces(file file.File) ([]Meta, error) {
	path := mustPathString("$.spec.workspaces[*]")
	var params []Workspace
	err := path.Read(strings.NewReader(string(file)), &params)
	var meta []Meta
	for _, p := range params {
		meta = append(meta, p)
	}
	return meta, err
}

func (f *File) parseWorkspaces() error {
	var workspaces []Workspace
	err := f.readPath(workspacesPath, &workspaces)
	if err != nil {
		return err
	}
	var meta []Meta
	for _, p := range workspaces {
		meta = append(meta, p)
	}
	f.workspaces = meta
	return nil
}
