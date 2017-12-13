// Copyright (C) 2017. See AUTHORS.

package files

import (
	"context"
	"encoding/json"

	"github.com/spacemonkeygo/rothko/disk"
	"github.com/zeebo/errs"
)

// filesMaker makes a *DB from the config.
func filesMaker(ctx context.Context, config string) (disk.Disk, error) {
	var conf struct {
		Options
		Dir string
	}

	if err := json.Unmarshal([]byte(config), &conf); err != nil {
		return nil, errs.Wrap(err)
	}

	return New(conf.Dir, conf.Options), nil
}
