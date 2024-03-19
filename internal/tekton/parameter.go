package tekton

import (
	"fmt"
)

type Parameter StringMap

var _ Meta = Parameter{}

func (p Parameter) Completions() []completion {
	return []completion{
		{
			text: fmt.Sprintf("$(params.%s)", p.Name()),
		},
	}
}

func (p Parameter) Name() string {
	n, _ := StringMap(p)["name"].(string)
	return n
}

func (p Parameter) Default() string {
	d, _ := StringMap(p)["default"].(string)
	return d
}

func (p Parameter) Type() string {
	if t, ok := StringMap(p)["type"].(string); ok {
		return t
	}
	return "string"
}

func (p Parameter) Description() string {
	d, _ := StringMap(p)["description"].(string)
	return d
}

func (p Parameter) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\ndefault: %s\ntype: %s\n%s\n```",
		p.Name(),
		p.Default(),
		p.Type(),
		p.Description(),
	)
}
