package file

import (
	"strings"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// TextDocument provides operations on a string representing the contents
// of a file.
type TextDocument string

// GetLine returns the string corresponding to the given line in the text.
func (f TextDocument) GetLine(line uint32) string {
	// TODO: improve this, text editor lib?
	lines := strings.Split(string(f), "\n")
	if line >= uint32(len(lines)) {
		return ""
	}
	return lines[line]
}

func (f TextDocument) Bytes() []byte {
	return []byte(string(f))
}

// FindPrevious finds the first position to the left of `pos` which contains
// any of the characters in `c`
func (f TextDocument) FindPrevious(c string, pos protocol.Position) int {
	line := f.GetLine(pos.Line)
	n := min(len(line), int(pos.Character))
	return strings.LastIndexAny(line[0:n], c)
}

// OffsetPosition returns the position in the file corresponding to the given
// offset.
func (f TextDocument) OffsetPosition(offset int) protocol.Position {
	// TODO: optimize
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

// LineOffset returns the offset corresponding to the first character of
// the given line.
func (f TextDocument) LineOffset(line int) int {
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

// PositionOffset returns the offset corresponding to the given position.
func (f TextDocument) PositionOffset(pos protocol.Position) int {
	return f.LineOffset(int(pos.Line)) + int(pos.Character)
}
