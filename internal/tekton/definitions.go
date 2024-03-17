package tekton

import protocol "github.com/tliron/glsp/protocol_3_16"

func (f *File) Definition(pos protocol.Position) *protocol.Range {
	ref := f.findDefinition(pos)
	if ref == nil || ref.ident == nil {
		return nil
	}
	return &ref.ident.prange
}

func (f *File) Hover(pos protocol.Position) *string {
	ref := f.findDefinition(pos)
	if ref == nil || ref.ident == nil {
		return nil
	}
	doc := ref.ident.meta.Documentation()
	return &doc
}
