package tekton

import (
	"github.com/goccy/go-yaml"
)

type StringMap = map[string]string

type Meta interface {
	Name() string
	Documentation() string
}

func mustPathString(path string) *yaml.Path {
	p, err := yaml.PathString(path)
	if err != nil {
		panic(err)
	}
	return p
}
