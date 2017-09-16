// Copyright (C) 2017. See AUTHORS.

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
