package completion

import (
	"fmt"
	"strings"
)

// Solve filters relevant completion suggestion given a query. A query is
// usually the word partially typed before a completion has been requested.
func Solve(query string, candidates []fmt.Stringer) []fmt.Stringer {
	rs := []fmt.Stringer{}
	for _, c := range candidates {
		// TODO: match subsequence instead of a prefix (fuzzy finder)
		if !strings.HasPrefix(c.String(), query) {
			continue
		}
		rs = append(rs, c)
	}
	return rs
}
