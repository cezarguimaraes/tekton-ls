package tekton

import protocol "github.com/tliron/glsp/protocol_3_16"

func (f *File) FindReferences(pos protocol.Position) []protocol.Range {
	// TODO
	return nil
}

func (d *Document) findReferences(pos protocol.Position) []protocol.Range {
	for _, id := range d.identifiers {
		if inRange(pos, id.prange) {
			return id.references
		}
	}
	return nil
}

func cmpPos(a, b protocol.Position) bool {
	if a.Line < b.Line {
		return true
	}
	if a.Line == b.Line && a.Character < b.Character {
		return true
	}
	return false
}

func inRange(pos protocol.Position, r protocol.Range) bool {
	// r.start <= pos && pos <= r.end
	return !cmpPos(pos, r.Start) && !cmpPos(r.End, pos)
}
