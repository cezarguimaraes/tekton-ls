package tekton

import (
	"os"
	"testing"
)

func TestWorkspace(t *testing.T) {
	w := NewWorkspace()
	cwd, _ := os.Getwd()
	folder := "file://" + cwd + "/testdata/workspace"
	w.AddFolder(folder)
	w.Lint()

	pipeURI := folder + "/pipe.yaml"
	taskURI := folder + "/task.yaml"

	pipeFile := w.File(pipeURI)
	if pipeFile == nil {
		t.Fatalf("expected %q to be in the workspace", pipeURI)
	}

	taskFile := w.File(taskURI)
	if taskFile == nil {
		t.Fatalf("expected %q to be in the workspace", taskURI)
	}

	identTCs := []struct {
		kind identifierKind
		name string
	}{
		{
			kind: IdentKindTask,
			name: "gen-code",
		},
	}

	for _, tc := range identTCs {
		id := w.getIdent(tc.kind, tc.name)
		if id == nil {
			t.Errorf("expected to find ident %q of kind %q in the workspace",
				tc.name, tc.kind,
			)
		}
	}
}
