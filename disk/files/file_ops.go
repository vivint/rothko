// Copyright (C) 2017. See AUTHORS.

package files

import "context"

// lastRecord returns the index of the last record in the file, using the
// metadata in the file as a hint.
func lastRecord(ctx context.Context, f file) (n int, err error) {
	m, err := f.Metadata(ctx)
	if err != nil {
		return 0, err
	}

	head := m.Head
	if head >= f.Capacity() {
		return head, nil
	}

	// figure out if head has a record. typically, it won't.
	orig_has := f.HasRecord(ctx, head)

	// if we have a record, walk forward. otherwise, we need to walk backward.
	delta := 1
	if !orig_has {
		delta = -1
	}

	// walk until we find something different from what we had. if we were
	// walking backwards, we will have to return head + 1.
	for {
		head += delta

		if head >= f.Capacity() {
			return head, nil
		}
		if head < 0 {
			return 0, nil
		}

		has := f.HasRecord(ctx, head)
		if has != orig_has {
			if orig_has {
				return head, nil
			}
			return head + 1, nil
		}
	}
}
