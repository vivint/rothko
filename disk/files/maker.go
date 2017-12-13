// Copyright (C) 2017. See AUTHORS.

package files

import (
	"context"

	"github.com/spacemonkeygo/rothko/disk"
	"github.com/zeebo/errs"
)

// filesMaker makes a *DB from the config.
func filesMaker(ctx context.Context, config string) (disk.Disk, error) {
	return nil, errs.New("unimplemented")
}
