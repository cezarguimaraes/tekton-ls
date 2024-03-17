package tekton

import (
	"fmt"
	"strings"

	"github.com/cezarguimaraes/tekton-lsp/internal/file"
)

type Result StringMap

func (p Result) Name() string {
	return StringMap(p)["name"]
}

func (p Result) Description() string {
	return StringMap(p)["description"]
}

func (p Result) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\n%s\n```",
		p.Name(),
		p.Description(),
	)
}

func Results(file file.File) ([]Meta, error) {
	path := mustPathString("$.spec.results[*]")
	var results []Result
	err := path.Read(strings.NewReader(string(file)), &results)
	var meta []Meta
	for _, p := range results {
		meta = append(meta, p)
	}
	return meta, err
}
