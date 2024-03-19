package tekton

type PipelineTask StringMap

var _ Meta = PipelineTask{}

func (p PipelineTask) Completions() []completion {
	return []completion{
		{
			text:    p.Name(),
			context: mustPathString("$.spec.tasks[*].runAfter"),
		},
	}
}

func (p PipelineTask) Name() string {
	n, _ := StringMap(p)["name"].(string)
	return n
}

func (p PipelineTask) Documentation() string {
	return ""
}
