package tekton

type Task StringMap

var _ Meta = PipelineTask{}

func (p Task) Completions() []completion {
	return []completion{
		{},
	}
}

func (p Task) Name() string {
	meta, ok := StringMap(p)["metadata"].(map[string]interface{})
	if !ok {
		return ""
	}
	n, _ := meta["name"].(string)
	return n
}

func (p Task) Documentation() string {
	return ""
}
