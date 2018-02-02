// Copyright (C) 2018. See AUTHORS.

package query

// glob matches a pattern with the name, where the name is allowed to be longer
// than the pattern so that "a" matches "abcd". The pattern should only use
// lower case alphanumeric strings, and the name will be matched as if it was
// lowered.
func glob(pattern, name string) bool {
	px, nx := 0, 0
	next_px, next_nx := 0, 0

	for px < len(pattern) {
		if nx >= len(name) {
			return false
		}
		n := name[nx]
		if 'A' <= n && n <= 'Z' {
			n += 'a' - 'A'
		}

		switch c := pattern[px]; c {
		default:
			if n == c {
				px++
				nx++
				continue
			}

		case '?':
			px++
			nx++
			continue

		case '*':
			next_px = px
			next_nx = nx + 1
			px++
			continue
		}

		if 0 < next_nx && next_nx < len(name) {
			px = next_px
			nx = next_nx
			continue
		}

		return false
	}

	return true
}
