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
	orig_has, err := f.HasRecord(ctx, head)
	if err != nil {
		return 0, err
	}

	// if we have a record, walk forward.
	delta := 1
	if !orig_has {
		delta = -1
	}

	// walk until we find something different from what we have. if we were
	// walking backwards, we will have to return head + 1.
	for {
		head += delta

		if head >= f.Capacity() {
			return head, nil
		}
		has, err := f.HasRecord(ctx, head)
		if err != nil {
			return 0, err
		}

		if has != orig_has {
			if orig_has {
				return head, nil
			}
			return head + 1, nil
		}
	}
}
