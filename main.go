package main

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
	"github.com/tliron/glsp/server"

	// Must include a backend implementation
	// See CommonLog for other options: https://github.com/tliron/commonlog
	_ "github.com/tliron/commonlog/simple"
)

const lsName = "tekton"

var (
	version string = "0.0.1"
	handler protocol.Handler
	srv     *server.Server
	files   map[string]file.File = make(map[string]file.File)
)

func main() {
	// This increases logging verbosity (optional)
	commonlog.Configure(2, nil)

	handler = protocol.Handler{
		Initialize:             initialize,
		Initialized:            initialized,
		Shutdown:               shutdown,
		SetTrace:               setTrace,
		TextDocumentHover:      hover,
		TextDocumentDidOpen:    docOpen,
		TextDocumentDidChange:  docChange,
		TextDocumentCompletion: docCompletion,
	}

	srv = server.NewServer(&handler, lsName, true)

	srv.RunStdio()
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()
	value := protocol.TextDocumentSyncKindFull
	capabilities.TextDocumentSync.(*protocol.TextDocumentSyncOptions).Change = &value
	capabilities.CompletionProvider.TriggerCharacters = []string{
		".",
		"(",
	}

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}, nil
}

func publishDiagnostics(context *glsp.Context, doc protocol.VersionedTextDocumentIdentifier) error {
	dgs, err := tekton.Diagnostics(srv.Log, files[doc.URI])
	if err != nil {
		return err
	}

	ver := uint32(doc.Version)
	context.Notify(
		protocol.ServerTextDocumentPublishDiagnostics,
		protocol.PublishDiagnosticsParams{
			URI: doc.URI,
			// TODO: match .Version and params.TextDocument.Version types
			Diagnostics: dgs,
			Version:     &ver,
		},
	)
	return nil
}

func docOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	files[params.TextDocument.URI] = file.File(params.TextDocument.Text)
	return publishDiagnostics(context, protocol.VersionedTextDocumentIdentifier{
		TextDocumentIdentifier: protocol.TextDocumentIdentifier{
			URI: params.TextDocument.URI,
		},
		Version: params.TextDocument.Version,
	})
}

func docChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) != 1 {
		return fmt.Errorf("expected event to contain a single ContentChange")
	}
	files[params.TextDocument.URI] = file.File(
		params.ContentChanges[0].(protocol.TextDocumentContentChangeEventWhole).Text,
	)
	return publishDiagnostics(context, params.TextDocument)
}

func docCompletion(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	var cs []protocol.CompletionItem
	f := files[params.TextDocument.URI]

	start := f.FindPrevious("$", params.Position)
	if start != -1 {
		line := f.GetLine(params.Position.Line)
		query := line[start:min(len(line), int(params.Position.Character))]

		candidates := tekton.Completions(srv.Log, f)

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

func hover(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	f, err := parser.ParseBytes([]byte(files[params.TextDocument.URI]), parser.ParseComments)
	if err != nil {
		return nil, err
	}

	node := yaml.FindNode(f.Docs[0], int(params.Position.Line)+1, int(params.Position.Character)+1)
	srv.Log.Debug("hover pos", "line", int(params.Position.Line)+1, "character", int(params.Position.Character)+1)
	s := "<null>"
	if node != nil {
		s = node.String()
		srv.Log.Info("anao")
	}

	srv.Log.Info("hover called")
	hv := &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: fmt.Sprintf("test\n```yaml\n%s\n```", s),
		},
	}
	return hv, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
