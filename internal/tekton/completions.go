package tekton

import (
	"fmt"
)

type CompletionCandidate struct {
	Text  string
	Value Meta
}

func (c CompletionCandidate) String() string {
	return c.Text
}

func (d *Document) completions() []fmt.Stringer {
	res := []fmt.Stringer{}

	for _, id := range d.identifiers {
		cs := id.meta.Completions()
		for _, c := range cs {
			res = append(res, CompletionCandidate{
				Text:  c,
				Value: id.meta,
			})
		}
	}

	return res
}
