package tekton

import (
	"fmt"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

type CompletionCandidate struct {
	Text  string
	Value Meta
}

func (c CompletionCandidate) String() string {
	return c.Text
}

func (f *File) Completions(pos protocol.Position) []fmt.Stringer {
	res := []fmt.Stringer{}
	if f.parseError != nil {
		return res
	}
	// TODO: find doc for pos
	return res
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
