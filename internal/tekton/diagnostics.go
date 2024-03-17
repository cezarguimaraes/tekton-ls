package tekton

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/cezarguimaraes/tekton-lsp/internal/file"
	"github.com/goccy/go-yaml"
	"github.com/tliron/commonlog"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type diag struct {
	label    string
	regexp   *regexp.Regexp
	listFunc func(file.File) ([]Meta, error)
}

var diags = []diag{
	{
		label:    "parameter",
		regexp:   regexp.MustCompile(`\$\(params\.(.*?)\)`),
		listFunc: Parameters,
	},
	{
		label:    "result",
		regexp:   regexp.MustCompile(`\$\(results\.(.*?)\.(.*?)\)`),
		listFunc: Results,
	},
}

func Diagnostics(log commonlog.Logger, file file.File) ([]protocol.Diagnostic, error) {
	rs := make([]protocol.Diagnostic, 0)
	for _, diag := range diags {
		ls, err := diag.listFunc(file)
		if err != nil && !errors.Is(err, yaml.ErrNotFoundNode) {
			log.Error("error listing references", "label", diag.label, "error", err)
			continue
		}
		m := mapFromSlice(ls)

		refs := diag.regexp.FindAllSubmatchIndex([]byte(file), 1000)
		sev := protocol.DiagnosticSeverityError
		src := "validation"

		for _, match := range refs {
			name := string(file)[match[2]:match[3]]
			if _, ok := m[name]; ok {
				continue
			}
			rs = append(rs, protocol.Diagnostic{
				Range: protocol.Range{
					Start: file.OffsetPosition(match[0]),
					End:   file.OffsetPosition(match[1]),
				},
				Message:  fmt.Sprintf("unknown %s %s", diag.label, name),
				Severity: &sev,
				Source:   &src,
			})
		}
	}
	return rs, nil
}

func mapFromSlice(s []Meta) map[string]struct{} {
	m := make(map[string]struct{})
	for _, p := range s {
		m[p.Name()] = struct{}{}
	}
	return m
}
