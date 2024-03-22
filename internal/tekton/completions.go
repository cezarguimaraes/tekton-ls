package tekton

import (
	"fmt"

	yaml_helper "github.com/cezarguimaraes/tekton-ls/internal/yaml"
	"github.com/goccy/go-yaml"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// completion defines a contextual completion. That is, a text to be suggested,
// and a yaml path required for it to be valid.
type completion struct {
	context *yaml.Path
	text    string
}

// CompletionCandidate holds the text of a completion and the Tekton object
// to which it refers.
type CompletionCandidate struct {
	// Text is the completion to be suggested.
	Text string

	// Value is the Tekton object to which the completion refers.
	Value Meta
}

// String implements fmt.Stringer for CompletionCandidate.
func (c CompletionCandidate) String() string {
	return c.Text
}

// Completion returns a list of completion suggestion given a position in
// a Tekton Document.
func (d *Document) completions(pos protocol.Position) []fmt.Stringer {
	res := []fmt.Stringer{}

	for _, id := range d.identifiers {
		cs := id.meta.Completions()
		for _, c := range cs {
			if c.context != nil {
				// contextual completion
				ctx, err := c.context.FilterNode(d.ast.Body)
				if err != nil {
					continue
				}
				if ctx == nil {
					continue
				}
				pos := yaml_helper.FindNode(ctx, int(pos.Line)+1, int(pos.Character)+1)
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
