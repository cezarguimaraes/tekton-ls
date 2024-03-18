package tekton

import (
	"fmt"

	"github.com/cezarguimaraes/tekton-lsp/internal/yaml"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type CompletionCandidate struct {
	Text  string
	Value Meta
}

func (c CompletionCandidate) String() string {
	return c.Text
}

func (d *Document) completions(pos protocol.Position) []fmt.Stringer {
	res := []fmt.Stringer{}

	for _, id := range d.identifiers {
		cs := id.meta.Completions()
		for _, c := range cs {
			if c.context != nil {
				// contextual completion
				ctx, _ := c.context.FilterNode(d.ast.Body)
				if ctx == nil {
					continue
				}
				pos := yaml.FindNode(ctx, int(pos.Line)+1, int(pos.Character)+1)
				if pos == nil {
					continue
				}
			}
			res = append(res, CompletionCandidate{
				Text:  c.text,
				Value: id.meta,
			})
		}
	}

	return res
}
