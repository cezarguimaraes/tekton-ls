package tekton

import protocol "github.com/tliron/glsp/protocol_3_16"

// definition returns the location of the definition of the identifier
// in a given position, or nil if none is found in this document.
func (d *Document) definition(pos protocol.Position) *protocol.Location {
	ref := d.referenceInPosition(pos)
	if ref == nil || ref.ident == nil {
		return nil
	}
	return &ref.ident.location
}

// hover returns the Tekton object documentation for the definition
// of the identifier in a given position, or nil if none is found.
func (d *Document) hover(pos protocol.Position) *string {
	ref := d.referenceInPosition(pos)
	if ref == nil || ref.ident == nil {
		return nil
	}
	doc := ref.ident.meta.Documentation()
	return &doc
}

// referenceInPosition searches for a reference in the document in the given
// position. A reference is any text fragment which might refer to an identifier.
func (d *Document) referenceInPosition(pos protocol.Position) *reference {
	for _, ref := range d.references {
		// assuming ref.start.Line = ref.end.Line
		if ref.start.Line != pos.Line {
			continue
		}
		if pos.Character > ref.end.Character {
			continue
		}
		if pos.Character < ref.start.Character {
			continue
		}
		return &ref
	}
	return nil
}
