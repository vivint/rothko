// Copyright (C) 2017. See AUTHORS.

package confparse

import (
	"strings"
)

// Parse takes comma separated key=value pairs and places them in a map.
func Parse(config string) map[string]string {
	out := make(map[string]string)
	for _, entry := range strings.Split(config, ",") {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 1 {
			out[strings.TrimSpace(entry)] = ""
		} else {
			out[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return out
}
