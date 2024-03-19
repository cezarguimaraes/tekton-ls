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
	return StringMap(p)["name"].(string)
}

func (p PipelineTask) Documentation() string {
	return ""
}
