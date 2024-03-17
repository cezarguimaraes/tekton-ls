package tekton

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/tliron/commonlog"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type diag struct {
	label    string
	regexp   *regexp.Regexp
	listFunc func(File) []Meta
}

var diags = []diag{
	{
		// TODO: validate object params
		label:    "parameter",
		regexp:   regexp.MustCompile(`\$\(params\.(.*?)\)`),
		listFunc: func(f File) []Meta { return f.parameters },
	},
	{
		// TODO: validate .path
		label:    "result",
		regexp:   regexp.MustCompile(`\$\(results\.(.*?)\.(.*?)\)`),
		listFunc: func(f File) []Meta { return f.results },
	},
	{
		// TODO: validate .path
		label:    "workspace",
		regexp:   regexp.MustCompile(`\$\(workspaces\.(.*?)\.(.*?)\)`),
		listFunc: func(f File) []Meta { return f.workspaces },
	},
}

func (f File) Diagnostics(log commonlog.Logger) ([]protocol.Diagnostic, error) {
	rs := make([]protocol.Diagnostic, 0)

	if f.parseError != nil {
		if d := syntaxErrorDiagnostic(f.parseError); d != nil {
			rs = append(rs, *d)
			return rs, nil
		}
	}

	for _, diag := range diags {
		ls := diag.listFunc(f)

		m := mapFromSlice(ls)

		refs := diag.regexp.FindAllSubmatchIndex(f.Bytes(), 1000)
		sev := protocol.DiagnosticSeverityError
		src := "validation"

		for _, match := range refs {
			name := string(f.Bytes())[match[2]:match[3]]
			if _, ok := m[name]; ok {
				continue
			}
			rs = append(rs, protocol.Diagnostic{
				Range: protocol.Range{
					Start: f.OffsetPosition(match[0]),
					End:   f.OffsetPosition(match[1]),
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
