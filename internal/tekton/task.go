package tekton

type IdentTask StringMap

var _ Meta = PipelineTask{}

func (p IdentTask) Completions() []completion {
	return []completion{}
}

func (p IdentTask) Name() string {
	meta, ok := StringMap(p)["metadata"].(map[string]interface{})
	if !ok {
		return ""
	}
	n, _ := meta["name"].(string)
	return n
}

func (p IdentTask) Documentation() string {
	return ""
}
