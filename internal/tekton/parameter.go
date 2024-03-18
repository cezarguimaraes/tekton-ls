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
	return StringMap(p)["name"]
}

func (p Parameter) Default() string {
	return StringMap(p)["default"]
}

func (p Parameter) Type() string {
	if t, ok := StringMap(p)["type"]; ok {
		return t
	}
	return "string"
}

func (p Parameter) Description() string {
	return StringMap(p)["description"]
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
