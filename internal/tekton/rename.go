package tekton

import (
	"fmt"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (d *Document) prepareRename(pos protocol.Position) *protocol.Range {
	id := d.findIdentifier(pos)
	if id == nil {
		return nil
	}
	return &id.prange
}

func (d *Document) rename(
	pos protocol.Position,
	newName string,
) (*protocol.WorkspaceEdit, error) {
	// maybe support renaming from references as well?
	id := d.findIdentifier(pos)
	if id == nil {
		return nil, fmt.Errorf("nothing to rename")
	}

	changes := map[string][]protocol.TextEdit{}
	changes[d.file.uri] = append(changes[d.file.uri],
		protocol.TextEdit{
			Range:   id.prange,
			NewText: newName,
		},
	)
	for _, ref := range id.references {
		changes[ref[1].URI] = append(changes[ref[1].URI],
			protocol.TextEdit{
				Range:   ref[1].Range,
				NewText: newName,
			},
		)
	}

	return &protocol.WorkspaceEdit{Changes: changes}, nil
}
