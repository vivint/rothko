// Copyright (C) 2017. See AUTHORS.

package files

import "context"

// getHeadPointer returns the index of the head pointer of the file, which is
// where the next write would go, using the metadata in the file as a hint.
func getHeadPointer(ctx context.Context, f file) (n int, err error) {
	m, err := f.Metadata(ctx)
	if err != nil {
		return 0, err
	}

	head := m.Head
	if head < 0 {
		return head, nil
	}

	// figure out if head has a record. typically, it won't.
	orig_has := f.HasRecord(ctx, head)

	// if we have a record, walk backward. otherwise, we need to walk forward.
	delta := -1
	if !orig_has {
		delta = 1
	}

	// walk until we find something different from what we had. if we are
	// walking backward, we just return the pointer. otherwise we need to
	// return head - 1.
	for {
		head += delta

		if head == f.Capacity() {
			return head - 1, nil
		}
		if head < 0 {
			return -1, nil
		}

		has := f.HasRecord(ctx, head)
		if has != orig_has {
			if orig_has {
				return head, nil
			}
			return head - 1, nil
		}
	}
}
