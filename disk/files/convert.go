// Copyright (C) 2017. See AUTHORS.

package files

// metricToPath takes a metric and converts it to a path on disk. It replaces
// dots with path separators, where when encountering a group of dots, only the
// first dot becomes a separator.
func metricToPath(metric string) string {
	// make room for .data as well as some escaping characters
	out := make([]byte, 0, len(metric)+5+(2*5))

	in_dot_group := false
	for i := 0; i < len(metric); i++ {
		switch ch := metric[i]; ch {
		case '.':
			if in_dot_group {
				out = append(out, '.')
			} else {
				out = append(out, '/')
				in_dot_group = true
			}
		case '/':
			in_dot_group = false
			out = append(out, '%', '2', 'f')
		case '%':
			in_dot_group = false
			out = append(out, '%', '2', '5')
		default:
			in_dot_group = false
			out = append(out, ch)
		}
	}

	out = append(out, ".data"...)
	return string(out)
}
