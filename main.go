package main

import (
	"github.com/cezarguimaraes/tekton-ls/internal/lsp"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp/server"

	// Must include a backend implementation
	// See CommonLog for other options: https://github.com/tliron/commonlog
	_ "github.com/tliron/commonlog/simple"
)

var version string = "dev"

func main() {
	// This increases logging verbosity (optional)
	commonlog.Configure(2, nil)

	th := lsp.NewTektonHandler(version)

	server := server.NewServer(&th.Handler, th.Name(), true)
	th.Log = server.Log

	server.RunStdio()
}
