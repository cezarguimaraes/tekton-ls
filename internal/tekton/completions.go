package tekton

import (
	"fmt"

	"github.com/tliron/commonlog"
)

type CompletionCandidate struct {
	Text  string
	Value Meta
}

func (c CompletionCandidate) String() string {
	return c.Text
}

func (f File) Completions(log commonlog.Logger) []fmt.Stringer {
	res := []fmt.Stringer{}
	if f.parseError != nil {
		return res
	}

	for _, id := range f.identifiers {
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
