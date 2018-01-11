// Copyright (C) 2017. See AUTHORS.

package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/spacemonkeygo/rothko/disk/files"
	"github.com/zeebo/errs"
)

func runRender(ctx context.Context, di *files.DB, metric string,
	dur time.Duration) error {

	now := time.Now()
	stop_before := now.Add(-dur).UnixNano()

	err := di.Query(ctx, metric, now.UnixNano(), nil,
		func(ctx context.Context, start, end int64, data []byte) (
			bool, error) {

			if end < stop_before {
				return true, nil
			}

			fmt.Println(start, end, hex.Dump(data))

			return false, nil
		})
	if err != nil {
		return errs.Wrap(err)
	}

	return nil
}
