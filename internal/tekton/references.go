package tekton

import (
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func wholeReferences(id *identifier) []protocol.Range {
	if id == nil {
		return nil
	}
	var refs []protocol.Range
	for _, ref := range id.references {
		refs = append(refs, ref[0])
	}
	return refs
}

func (d *Document) findIdentifier(pos protocol.Position) *identifier {
	for _, id := range d.identifiers {
		if inRange(pos, id.prange) {
			return id
		}
	}
	return nil
}

func (d *Document) findReferences(pos protocol.Position) []protocol.Range {
	return wholeReferences(d.findIdentifier(pos))
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
	// r.start <= pos && pos < r.end
	return !cmpPos(pos, r.Start) && cmpPos(pos, r.End)
}
