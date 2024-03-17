package tekton

import (
	"fmt"

	"github.com/tliron/commonlog"
)

type completion struct {
	listFunc func(File) []Meta
	format   string
}

var completions = []completion{
	{
		// TODO: object params
		listFunc: func(f File) []Meta { return f.parameters },
		format:   "$(params.%s)",
	},
	{
		listFunc: func(f File) []Meta { return f.results },
		format:   "$(results.%s.path)",
	},
	{
		listFunc: func(f File) []Meta { return f.workspaces },
		format:   "$(workspaces.%s.path)",
	},
}

type CompletionCandidate struct {
	Text  string
	Value Meta
}

func (c CompletionCandidate) String() string {
	return c.Text
}

func (f File) Completions(log commonlog.Logger) []fmt.Stringer {
	cs := []fmt.Stringer{}
	if f.parseError != nil {
		return cs
	}

	for _, compl := range completions {
		ls := compl.listFunc(f)
		for _, cand := range ls {
			cs = append(cs, CompletionCandidate{
				Text:  fmt.Sprintf(compl.format, cand.Name()),
				Value: cand,
			})
		}
	}
	return cs
}
