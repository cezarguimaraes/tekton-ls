package lsp

import (
	"fmt"

	"github.com/cezarguimaraes/tekton-lsp/internal/completion"
	"github.com/cezarguimaraes/tekton-lsp/internal/file"
	"github.com/cezarguimaraes/tekton-lsp/internal/tekton"
	"github.com/cezarguimaraes/tekton-lsp/internal/yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const (
	lsName  = "tekton-lsp"
	version = "0.0.1"
)

type TektonHandler struct {
	protocol.Handler

	Log commonlog.Logger

	files map[string]*tekton.File
}

func NewTektonHandler() *TektonHandler {
	th := &TektonHandler{
		files: make(map[string]*tekton.File),
	}
	th.Handler = protocol.Handler{
		Initialize:             th.initialize(),
		Initialized:            th.initialized(),
		Shutdown:               th.shutdown(),
		SetTrace:               th.setTrace(),
		TextDocumentHover:      th.hover(),
		TextDocumentDidOpen:    th.docOpen(),
		TextDocumentDidChange:  th.docChange(),
		TextDocumentCompletion: th.docCompletion(),
		TextDocumentDefinition: th.definition(),
	}
	return th
}

func (th *TektonHandler) Name() string {
	return lsName
}

type docTypes interface {
	protocol.TextDocumentIdentifier | protocol.VersionedTextDocumentIdentifier
}

func getDoc[T docTypes](th *TektonHandler, doc T) *tekton.File {
	var uri string
	switch d := any(doc).(type) {
	case protocol.TextDocumentIdentifier:
		uri = d.URI
	case protocol.VersionedTextDocumentIdentifier:
		uri = d.URI
	default:
		panic("unknown document identifier type")
	}
	return th.getDoc(uri)
}

func (th *TektonHandler) getDoc(uri string) *tekton.File {
	return th.files[uri]
}

func (th *TektonHandler) publishDiagnostics(context *glsp.Context, doc protocol.VersionedTextDocumentIdentifier) error {
	dgs, err := getDoc(th, doc).Diagnostics(th.Log)
	if err != nil {
		return err
	}

	ver := uint32(doc.Version)
	context.Notify(
		protocol.ServerTextDocumentPublishDiagnostics,
		protocol.PublishDiagnosticsParams{
			URI:         doc.URI,
			Diagnostics: dgs,
			Version:     &ver,
		},
	)
	return nil
}

func (th *TektonHandler) initialize() protocol.InitializeFunc {
	return func(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
		capabilities := th.Handler.CreateServerCapabilities()

		value := protocol.TextDocumentSyncKindFull
		capabilities.TextDocumentSync.(*protocol.TextDocumentSyncOptions).Change = &value

		capabilities.CompletionProvider.TriggerCharacters = []string{
			".",
			"(",
		}

		ver := version
		return protocol.InitializeResult{
			Capabilities: capabilities,
			ServerInfo: &protocol.InitializeResultServerInfo{
				Name:    lsName,
				Version: &ver,
			},
		}, nil
	}
}

func (th *TektonHandler) docOpen() protocol.TextDocumentDidOpenFunc {
	return func(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
		th.files[params.TextDocument.URI] = tekton.ParseFile(
			file.File(params.TextDocument.Text),
		)
		return th.publishDiagnostics(context, protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: params.TextDocument.URI,
			},
			Version: params.TextDocument.Version,
		})
	}
}

func (th *TektonHandler) docChange() protocol.TextDocumentDidChangeFunc {
	return func(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
		if len(params.ContentChanges) != 1 {
			return fmt.Errorf("expected event to contain a single ContentChange")
		}
		th.files[params.TextDocument.URI] = tekton.ParseFile(file.File(
			params.ContentChanges[0].(protocol.TextDocumentContentChangeEventWhole).Text,
		))
		return th.publishDiagnostics(context, params.TextDocument)
	}
}

func (th *TektonHandler) docCompletion() protocol.TextDocumentCompletionFunc {
	return func(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
		var cs []protocol.CompletionItem
		f := getDoc(th, params.TextDocument)

		start := f.FindPrevious("$", params.Position)
		if start != -1 {
			line := f.GetLine(params.Position.Line)
			query := line[start:min(len(line), int(params.Position.Character))]

			candidates := f.Completions(th.Log)

			matches := completion.Solve(query, candidates)
			kind := protocol.CompletionItemKindProperty
			for idx, m := range matches {
				preselect := idx == 0
				cs = append(cs, protocol.CompletionItem{
					Label:     m.String(),
					Kind:      &kind,
					Preselect: &preselect,
					Documentation: protocol.MarkupContent{
						Kind:  protocol.MarkupKindMarkdown,
						Value: m.(tekton.CompletionCandidate).Value.Documentation(),
					},
					TextEdit: protocol.TextEdit{
						NewText: m.String(),
						Range: protocol.Range{
							Start: protocol.Position{Line: params.Position.Line, Character: uint32(start)},
							End:   protocol.Position{Line: params.Position.Line, Character: params.Position.Character},
						},
					},
				})
			}
		}

		return cs, nil
	}
}

func (th *TektonHandler) definition() protocol.TextDocumentDefinitionFunc {
	return func(context *glsp.Context, params *protocol.DefinitionParams) (any, error) {
		f := getDoc(th, params.TextDocument)
		defPos := f.Definition(params.Position)
		if defPos == nil {
			return nil, nil
		}

		loc := protocol.Location{
			URI: params.TextDocument.URI,
			Range: protocol.Range{
				Start: *defPos,
				End:   *defPos,
			},
		}
		return loc, nil
	}
}

func (th *TektonHandler) hover() protocol.TextDocumentHoverFunc {
	return func(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
		f, err := parser.ParseBytes(getDoc(th, params.TextDocument).Bytes(), parser.ParseComments)
		if err != nil {
			return nil, err
		}

		node := yaml.FindNode(f.Docs[0], int(params.Position.Line)+1, int(params.Position.Character)+1)
		th.Log.Debug("hover pos", "line", int(params.Position.Line)+1, "character", int(params.Position.Character)+1)
		s := "<null>"
		if node != nil {
			s = node.String()
		}

		th.Log.Info("hover called")
		hv := &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: fmt.Sprintf("test\n```yaml\n%s\n```", s),
			},
		}
		return hv, nil
	}
}

func (th *TektonHandler) initialized() protocol.InitializedFunc {
	return func(context *glsp.Context, params *protocol.InitializedParams) error {
		return nil
	}
}

func (th *TektonHandler) shutdown() protocol.ShutdownFunc {
	return func(context *glsp.Context) error {
		protocol.SetTraceValue(protocol.TraceValueOff)
		return nil
	}
}

func (th *TektonHandler) setTrace() protocol.SetTraceFunc {
	return func(context *glsp.Context, params *protocol.SetTraceParams) error {
		protocol.SetTraceValue(params.Value)
		return nil
	}
}
