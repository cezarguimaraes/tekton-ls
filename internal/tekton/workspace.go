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
	return StringMap(p)["name"].(string)
}

func (p Workspace) Description() string {
	return StringMap(p)["description"].(string)
}

func (p Workspace) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\n%s\n```",
		p.Name(),
		p.Description(),
	)
}
