package tekton

import (
	"fmt"
)

type IdentResult StringMap

var _ Meta = IdentResult{}

func (p IdentResult) Completions() []completion {
	return []completion{
		{
			text: fmt.Sprintf("$(results.%s.path)", p.Name()),
		},
	}
}

func (p IdentResult) Name() string {
	n, _ := StringMap(p)["name"].(string)
	return n
}

func (p IdentResult) Description() string {
	d, _ := StringMap(p)["description"].(string)
	return d
}

func (p IdentResult) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\n%s\n```",
		p.Name(),
		p.Description(),
	)
}
