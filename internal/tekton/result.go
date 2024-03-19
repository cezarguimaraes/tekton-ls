package tekton

import (
	"fmt"
)

type Result StringMap

var _ Meta = Result{}

func (p Result) Completions() []completion {
	return []completion{
		{
			text: fmt.Sprintf("$(results.%s.path)", p.Name()),
		},
	}
}

func (p Result) Name() string {
	return StringMap(p)["name"].(string)
}

func (p Result) Description() string {
	return StringMap(p)["description"].(string)
}

func (p Result) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\n%s\n```",
		p.Name(),
		p.Description(),
	)
}
