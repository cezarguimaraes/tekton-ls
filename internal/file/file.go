package file

import (
	"strings"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// TODO: stop stuttering
type File string

// TODO: improve this, text editor lib?
func (f File) GetLine(line uint32) string {
	lines := strings.Split(string(f), "\n")
	if line >= uint32(len(lines)) {
		return ""
	}
	return lines[line]
}

func (f File) Bytes() []byte {
	return []byte(string(f))
}

func (f File) getContext(pos protocol.Position, c int) string {
	line := f.GetLine(pos.Line)
	from := max(0, int(pos.Character)-c)
	return line[from:pos.Character]
}

func (f File) HasContext(pos protocol.Position, ctx string) bool {
	return f.getContext(pos, len(ctx)) == ctx
}

func (f File) FindPrevious(c string, pos protocol.Position) int {
	line := f.GetLine(pos.Line)
	n := min(len(line), int(pos.Character))
	return strings.LastIndex(line[0:n], c)
}

// TODO: optimize
func (f File) OffsetPosition(offset int) protocol.Position {
	s := string(f)
	line := uint32(0)
	column := uint32(0)
	for i := range offset {
		column++
		if s[i] == '\n' {
			column = 0
			line += 1
		}
	}
	return protocol.Position{
		Line:      line,
		Character: column,
	}
}

func (f File) LineOffset(line int) int {
	s := []byte(string(f))
	offset := 0
	for line > 0 && offset < len(s) {
		if s[offset] == '\n' {
			line--
		}
		offset++
	}
	return offset
}
