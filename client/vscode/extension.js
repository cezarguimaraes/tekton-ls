const { LanguageClient, TransportKind } = require("vscode-languageclient/node");

module.exports = {
    /** @param {import("vscode").ExtensionContext} context*/
    activate(context) {
        /** @type {import("vscode-languageclient/node").ServerOptions} */
        const serverOptions = {
            run: {
                command: "tekton-ls",
                transport: TransportKind.stdio
            },
            debug: {
                command: "tekton-ls",
                transport: TransportKind.stdio
            },
        };

        /** @type {import("vscode-languageclient/node").LanguageClientOptions} */
        const clientOptions = {
            documentSelector: [{ scheme: "file", language: "yaml" }],
        };

        const client = new LanguageClient(
            "tekton-ls",
            "Tekton Pipelines Language Server",
            serverOptions,
            clientOptions
        );

        client.start();
    },
};
