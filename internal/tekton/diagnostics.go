package tekton

import (
	"fmt"
	"regexp"
	"strconv"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (d *Document) diagnostics() []protocol.Diagnostic {
	rs := make([]protocol.Diagnostic, 0)

	for _, ref := range d.references {
		if ref.ident != nil {
			continue
		}

		sev := protocol.DiagnosticSeverityError
		src := "validation"

		rs = append(rs, protocol.Diagnostic{
			Range: protocol.Range{
				Start: d.OffsetPosition(ref.offsets[0]),
				End:   d.OffsetPosition(ref.offsets[1]),
			},
			Message:  fmt.Sprintf("unknown %s %s", ref.kind, ref.name),
			Severity: &sev,
			Source:   &src,
		})
	}

	return rs
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
