// Copyright (C) 2018. See AUTHORS.

package files

import (
	"context"

	"github.com/spacemonkeygo/rothko/database"
	"github.com/spacemonkeygo/rothko/internal/typeassert"
	"github.com/spacemonkeygo/rothko/registry"
)

func init() {
	registry.RegisterDatabase("files", registry.DatabaseMakerFunc(
		func(ctx context.Context, config interface{}) (database.DB, error) {
			a := typeassert.A(config)
			dir := a.I("directory").String()
			opts := Options{
				Size:  a.I("size").Int(),
				Cap:   a.I("cap").Int(),
				Files: a.I("files").Int(),
				Tuning: Tuning{
					Buffer:  a.I("tuning").I("buffer").Int(),
					Drop:    a.I("tuning").I("drop").Bool(),
					Handles: a.I("tuning").I("handles").Int(),
					Workers: a.I("tuning").I("workers").Int(),
				},
			}
			if err := a.Err(); err != nil {
				return nil, err
			}

			return New(dir, opts), nil
		}))
}
