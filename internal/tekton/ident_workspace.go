package tekton

import (
	"fmt"
)

type IdentWorkspace StringMap

var _ Meta = IdentWorkspace{}

func (p IdentWorkspace) Completions() []completion {
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

func (p IdentWorkspace) Name() string {
	n, _ := StringMap(p)["name"].(string)
	return n
}

func (p IdentWorkspace) Description() string {
	d, _ := StringMap(p)["description"].(string)
	return d
}

func (p IdentWorkspace) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\n%s\n```",
		p.Name(),
		p.Description(),
	)
}
