package tekton

import (
	"fmt"
	"strings"

	"github.com/cezarguimaraes/tekton-lsp/internal/file"
)

var parametersPath = mustPathString("$.spec.parameters[*]")

type Parameter StringMap

func (p Parameter) Name() string {
	return StringMap(p)["name"]
}

func (p Parameter) Default() string {
	return StringMap(p)["default"]
}

func (p Parameter) Type() string {
	if t, ok := StringMap(p)["type"]; ok {
		return t
	}
	return "string"
}

func (p Parameter) Description() string {
	return StringMap(p)["description"]
}

func (p Parameter) Documentation() string {
	return fmt.Sprintf(
		"```yaml\nname: %s\ndefault: %s\ntype: %s\n%s\n```",
		p.Name(),
		p.Default(),
		p.Type(),
		p.Description(),
	)
}

func (f *File) parseParameters() error {
	var params []Parameter
	err := f.readPath(parametersPath, &params)
	if err != nil {
		return err
	}
	var meta []Meta
	for _, p := range params {
		meta = append(meta, p)
	}
	f.parameters = meta
	return nil
}

// TODO: pre-parse file
func Parameters(file file.File) ([]Meta, error) {
	path := mustPathString("$.spec.parameters[*]")
	var params []Parameter
	err := path.Read(strings.NewReader(string(file)), &params)
	var meta []Meta
	for _, p := range params {
		meta = append(meta, p)
	}
	return meta, err
}
