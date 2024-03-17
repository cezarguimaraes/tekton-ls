package tekton

import (
	"fmt"
)

type Workspace StringMap

func (p Workspace) Completions() []string {
	return []string{fmt.Sprintf("$(workspaces.%s.path)", p.Name())}
}

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
