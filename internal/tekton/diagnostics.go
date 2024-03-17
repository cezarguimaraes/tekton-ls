package tekton

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

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
		// TODO: validate object params
		label:    "parameter",
		regexp:   regexp.MustCompile(`\$\(params\.(.*?)\)`),
		listFunc: Parameters,
	},
	{
		// TODO: validate .path
		label:    "result",
		regexp:   regexp.MustCompile(`\$\(results\.(.*?)\.(.*?)\)`),
		listFunc: Results,
	},
	{
		// TODO: validate .path
		label:    "workspaces",
		regexp:   regexp.MustCompile(`\$\(workspaces\.(.*?)\.(.*?)\)`),
		listFunc: Workspaces,
	},
}

func Diagnostics(log commonlog.Logger, file file.File) ([]protocol.Diagnostic, error) {
	rs := make([]protocol.Diagnostic, 0)
	for _, diag := range diags {
		ls, err := diag.listFunc(file)

		if err != nil && !errors.Is(err, yaml.ErrNotFoundNode) {
			if d := syntaxErrorDiagnostic(err); d != nil {
				rs = append(rs, *d)
				break
			}
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

var syntaxErrorRegexp = regexp.MustCompile(`(?s)^\[(\d+):(\d+)\] (.+)`)

// hack to extract error position from goccy/go-yaml unexported syntaxError
func syntaxErrorDiagnostic(err error) *protocol.Diagnostic {
	ms := syntaxErrorRegexp.FindStringSubmatch(err.Error())
	if len(ms) != 4 {
		return nil
	}
	line, err := strconv.Atoi(ms[1])
	if err != nil {
		return nil
	}
	col, err := strconv.Atoi(ms[2])
	if err != nil {
		return nil
	}
	pos := protocol.Position{
		Line:      uint32(line - 1),
		Character: uint32(col - 1),
	}
	sev := protocol.DiagnosticSeverityError
	src := "syntax"
	return &protocol.Diagnostic{
		Range: protocol.Range{
			Start: pos,
			End:   pos,
		},
		Message:  ms[3],
		Severity: &sev,
		Source:   &src,
	}
}

func mapFromSlice(s []Meta) map[string]struct{} {
	m := make(map[string]struct{})
	for _, p := range s {
		m[p.Name()] = struct{}{}
	}
	return m
}
