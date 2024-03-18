package tekton

import (
	"fmt"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (d *Document) prepareRename(pos protocol.Position) bool {
	return d.findIdentifier(pos) != nil
}

func (d *Document) rename(
	pos protocol.Position,
	newName string,
) ([]protocol.TextEdit, error) {
	// maybe support renaming from references as well?
	id := d.findIdentifier(pos)
	if id == nil {
		return nil, fmt.Errorf("nothing to rename")
	}

	es := []protocol.TextEdit{
		{
			Range:   id.prange,
			NewText: newName,
		},
	}
	for _, ref := range id.references {
		es = append(es, protocol.TextEdit{
			Range:   ref[1],
			NewText: newName,
		})
	}

	return es, nil
}
