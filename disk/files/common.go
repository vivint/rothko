// Copyright (C) 2017. See AUTHORS.

package files

// fastMod computes n % m assuming that n is a random number in the full
// uint32 range.
func fastMod(n uint32, m int) int {
	return int((uint64(n) * uint64(m)) >> 32)
}

// copyStringSet makes a copy of a string set.
func copyStringSet(x map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(x))
	for key := range x {
		out[key] = struct{}{}
	}
	return out
}
