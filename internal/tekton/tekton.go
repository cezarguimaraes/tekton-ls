package tekton

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cezarguimaraes/tekton-lsp/internal/file"
	"github.com/goccy/go-yaml"
)

var paramRegexp *regexp.Regexp = regexp.MustCompile(`\$\(params\.(.*?)\)`)

type StringMap = map[string]string

type (
	Parameter StringMap
	Result    StringMap
)

type Meta interface {
	Name() string
	Documentation() string
}

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

func mustPathString(path string) *yaml.Path {
	p, err := yaml.PathString(path)
	if err != nil {
		panic(err)
	}
	return p
}
