package completion

import (
	"fmt"
	"strings"
)

func Solve(query string, candidates []fmt.Stringer) []fmt.Stringer {
	rs := []fmt.Stringer{}
	for _, c := range candidates {
		if !strings.HasPrefix(c.String(), query) {
			continue
		}
		rs = append(rs, c)
	}
	return rs
}
