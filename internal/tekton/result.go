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
	n, _ := StringMap(p)["name"].(string)
	return n
}

func (p Result) Description() string {
	d, _ := StringMap(p)["description"].(string)
	return d
}

func (p Result) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\n%s\n```",
		p.Name(),
		p.Description(),
	)
}
