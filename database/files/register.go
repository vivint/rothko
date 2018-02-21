// Copyright (C) 2018. See AUTHORS.

package files

import (
	"context"

	"github.com/vivint/rothko/database"
	"github.com/vivint/rothko/internal/typeassert"
	"github.com/vivint/rothko/registry"
)

func init() {
	registry.RegisterDatabase("files", registry.DatabaseMakerFunc(
		func(ctx context.Context, config interface{}) (database.DB, error) {
			a := typeassert.A(config)
			dir := a.I("directory").String()
			opts := Options{
				Size:  int(a.I("size").Int64()),
				Cap:   int(a.I("cap").Int64()),
				Files: int(a.I("files").Int64()),
				Tuning: Tuning{
					Buffer:  int(a.I("tuning").I("buffer").Int64()),
					Drop:    a.I("tuning").I("drop").Bool(),
					Handles: int(a.I("tuning").I("handles").Int64()),
					Workers: int(a.I("tuning").I("workers").Int64()),
				},
			}
			if err := a.Err(); err != nil {
				return nil, err
			}

			return New(dir, opts), nil
		}))
}
