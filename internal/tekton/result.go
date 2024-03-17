package tekton

import (
	"fmt"
)

type Result StringMap

func (p Result) Completions() []string {
	return []string{fmt.Sprintf("$(results.%s.path)", p.Name())}
}

func (p Result) Name() string {
	return StringMap(p)["name"]
}

func (p Result) Description() string {
	return StringMap(p)["description"]
}

func (p Result) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\n%s\n```",
		p.Name(),
		p.Description(),
	)
}
