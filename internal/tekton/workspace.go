package tekton

import (
	"fmt"
)

type Workspace StringMap

var _ Meta = Workspace{}

func (p Workspace) Completions() []completion {
	return []completion{
		{
			text: fmt.Sprintf("$(workspaces.%s.path)", p.Name()),
		},
		{
			text:    p.Name(),
			context: mustPathString("$.spec.tasks[*].workspaces[*].workspace"),
		},
	}
}

func (p Workspace) Name() string {
	n, _ := StringMap(p)["name"].(string)
	return n
}

func (p Workspace) Description() string {
	d, _ := StringMap(p)["description"].(string)
	return d
}

func (p Workspace) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\n%s\n```",
		p.Name(),
		p.Description(),
	)
}
