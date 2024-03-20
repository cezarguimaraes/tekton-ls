package tekton

import (
	"fmt"
)

type IdentParameter StringMap

var _ Meta = IdentParameter{}

func (p IdentParameter) Completions() []completion {
	return []completion{
		{
			text: fmt.Sprintf("$(params.%s)", p.Name()),
		},
	}
}

func (p IdentParameter) Name() string {
	n, _ := StringMap(p)["name"].(string)
	return n
}

func (p IdentParameter) Default() string {
	d, _ := StringMap(p)["default"].(string)
	return d
}

func (p IdentParameter) Type() string {
	if t, ok := StringMap(p)["type"].(string); ok {
		return t
	}
	return "string"
}

func (p IdentParameter) Description() string {
	d, _ := StringMap(p)["description"].(string)
	return d
}

func (p IdentParameter) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\ndefault: %s\ntype: %s\n%s\n```",
		p.Name(),
		p.Default(),
		p.Type(),
		p.Description(),
	)
}
