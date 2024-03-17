package tekton

import (
	"errors"
	"fmt"

	"github.com/cezarguimaraes/tekton-lsp/internal/file"
	"github.com/goccy/go-yaml"
	"github.com/tliron/commonlog"
)

type completion struct {
	listFunc func(file.File) ([]Meta, error)
	format   string
}

var completions = []completion{
	{
		listFunc: Parameters,
		format:   "$(params.%s)",
	},
	{
		listFunc: Results,
		format:   "$(results.%s.path)",
	},
}

type CompletionCandidate struct {
	Text  string
	Value Meta
}

func (c CompletionCandidate) String() string {
	return c.Text
}

func Completions(log commonlog.Logger, file file.File) []fmt.Stringer {
	cs := []fmt.Stringer{}
	for _, compl := range completions {
		ls, err := compl.listFunc(file)
		if err != nil && !errors.Is(err, yaml.ErrNotFoundNode) {
			log.Error("error listing completions", "error", err)
			continue
		}
		for _, cand := range ls {
			cs = append(cs, CompletionCandidate{
				Text:  fmt.Sprintf(compl.format, cand.Name()),
				Value: cand,
			})
		}
	}
	return cs
}
