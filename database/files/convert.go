// Copyright (C) 2018. See AUTHORS.

package files

import "strconv"

// metricToDir takes a metric and converts it to a path on disk. It replaces
// dots with path separators, where when encountering a group of dots, only the
// first dot becomes a separator. Path characters are replaced with urlencoded
// values.
func metricToDir(buf []byte, metric string) []byte {
	non_dots := false
	dots := 0
	for i := 0; i < len(metric); i++ {
		switch ch := metric[i]; ch {
		case '.':
			if dots > 0 || !non_dots {
				buf = append(buf, "%2e"...)
			} else {
				buf = append(buf, '/')
			}
			dots++
		default:
			if !non_dots && dots > 0 {
				buf = append(buf, '/')
			}
			non_dots = true
			dots = 0
			switch ch {
			case '/':
				buf = append(buf, "%2f"...)
			case '%':
				buf = append(buf, "%25"...)
			default:
				buf = append(buf, ch)
			}
		}
	}

	// fix up if we ended on dots and there were other characters. the case of
	// all dots is handled by the above loop and requires no work.
	if dots > 0 && non_dots {
		back := 1 + (dots-1)*3
		buf = buf[:len(buf)-back]
		buf = append(buf, '/')
		for i := 0; i < dots; i++ {
			buf = append(buf, "%2e"...)
		}
	}

	return buf
}

func dirToMetric(buf, dir []byte) ([]byte, error) {
	last_dot := false
	all_dots := false
	for i := 0; i < len(dir); i++ {
		switch ch := dir[i]; ch {
		case '/':
			if !last_dot {
				buf = append(buf, '.')
				last_dot = true
				all_dots = true
			}
		case '%':
			i++
			if i >= len(dir) || dir[i] != '2' {
				return nil, Error.New("invalid dir: %q", dir)
			}
			i++
			if i >= len(dir) {
				return nil, Error.New("invalid dir: %q", dir)
			}
			switch ch := dir[i]; ch {
			case 'e':
				buf = append(buf, '.')
				last_dot = true
			case 'f':
				buf = append(buf, '/')
				last_dot = false
				all_dots = false
			case '5':
				buf = append(buf, '%')
				last_dot = false
				all_dots = false
			default:
				return nil, Error.New("invalid dir: %q", dir)
			}
		case '.':
			return nil, Error.New("invalid dir: %q", dir)
		default:
			buf = append(buf, ch)
			last_dot = false
			all_dots = false
		}
	}

	// in the case we had /%2e%2e%2e we need to strip one dot off the end
	if all_dots && len(dir) > 0 {
		buf = buf[:len(buf)-1]
	}

	return buf, nil
}

// metricToPath selects the data file number out of the directory identified by
// the metric with metricToDir.
func metricToPath(buf []byte, metric string, num int) []byte {
	buf = metricToDir(buf, metric)

	if len(buf) > 0 {
		buf = append(buf, '/')
	}

	buf = strconv.AppendInt(buf, int64(num), 10)
	buf = append(buf, ".data"...)

	return buf
}
