package tekton

import protocol "github.com/tliron/glsp/protocol_3_16"

func (f *File) Definition(pos protocol.Position) *protocol.Position {
	ref := f.findReference(pos)
	if ref == nil || ref.ident == nil {
		return nil
	}
	ipos := ref.ident.definition.GetToken().Position
	return &protocol.Position{
		Line:      uint32(ipos.Line - 1),
		Character: uint32(ipos.Column - 1),
	}
}
