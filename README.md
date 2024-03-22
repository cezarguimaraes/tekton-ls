# tekton-ls

`tekton-ls` is a work-in-progress language server for [Tekton Pipelines](https://github.com/tektoncd/pipeline).
It currently supports `auto-completion`, `go-to-definition`, `find-references`, `rename` and `diagnostics` for:

- Task and Pipeline parameters
- Task and Pipeline results
- Task and Pipeline Workpaces
- PipelineTasks
- Tasks

## Installing

### VSCode

1. Install the language server
    
    ```bash
    go install github.com/cezarguimaraes/tekton-ls
    ```
2. Download the packaged extension from [./client/vscode/tekton-ls-0.0.1.vsix](https://github.com/cezarguimaraes/tekton-ls/raw/main/client/vscode/tekton-ls-0.0.1.vsix)
    
    ```bash
    wget https://github.com/cezarguimaraes/tekton-ls/raw/main/client/vscode/tekton-ls-0.0.1.vsix
    ```
3. Open the command palette in VScode (Ctrl+Shift+P / Cmd+Shift+P)
4. Choose the option `Extensions: Install from VISIX`
5. Navigate to the folder you downloaded the packaged extension in step 1 and select `tekton-ls-0.0.1.vsix`
