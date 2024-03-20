package tekton

import protocol "github.com/tliron/glsp/protocol_3_16"

func (d *Document) definition(pos protocol.Position) *protocol.Location {
	ref := d.findDefinition(pos)
	if ref == nil || ref.ident == nil {
		return nil
	}
	return &ref.ident.location
}

func (d *Document) hover(pos protocol.Position) *string {
	ref := d.findDefinition(pos)
	if ref == nil || ref.ident == nil {
		return nil
	}
	doc := ref.ident.meta.Documentation()
	return &doc
}

func (d *Document) findDefinition(pos protocol.Position) *reference {
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
