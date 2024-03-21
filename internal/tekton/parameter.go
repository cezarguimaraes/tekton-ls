package tekton

import (
	"fmt"
	"strings"
)

type identParam struct {
	value      interface{}
	parent     interface{}
	parentKind string
	parentName string
}

func IdentParameter(v StringMap, parent interface{}) *identParam {
	p := &identParam{
		value:  v,
		parent: parent,
	}

	pm := parent.(map[string]interface{})
	kind, ok := pm["kind"].(string)
	if ok {
		p.parentKind = strings.ToLower(kind)
	}

	var meta map[string]interface{}
	if m, ok := pm["metadata"]; ok {
		meta, _ = m.(map[string]interface{})
	}

	if meta != nil {
		p.parentName, _ = meta["name"].(string)
	}

	return p
}

var _ Meta = &identParam{}

func (p *identParam) Completions() []completion {
	return []completion{
		{
			text: fmt.Sprintf("$(params.%s)", p.Name()),
		},
	}
}

func (p *identParam) Name() string {
	n, _ := StringMap(p.value.(map[string]interface{}))["name"].(string)
	return n
}

func (p *identParam) Default() string {
	d, _ := StringMap(p.value.(map[string]interface{}))["default"].(string)
	return d
}

func (p *identParam) Type() string {
	if t, ok := StringMap(p.value.(map[string]interface{}))["type"].(string); ok {
		return t
	}
	return "string"
}

func (p *identParam) Description() string {
	d, _ := StringMap(p.value.(map[string]interface{}))["description"].(string)
	return d
}

func (p *identParam) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\ndefault: %s\ntype: %s\n%s\n```",
		p.Name(),
		p.Default(),
		p.Type(),
		p.Description(),
	)
}
