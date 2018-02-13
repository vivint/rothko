// Copyright (C) 2018. See AUTHORS.

package tdigest

import (
	"context"

	"github.com/spacemonkeygo/rothko/dist"
	"github.com/spacemonkeygo/rothko/internal/typeassert"
	"github.com/spacemonkeygo/rothko/registry"
)

func init() {
	registry.RegisterDistribution("tdigest", registry.DistributionMakerFunc(
		func(ctx context.Context, config interface{}) (dist.Params, error) {
			a := typeassert.A(config)
			params := Params{
				Compression: a.I("compression").Float64(),
			}
			if err := a.Err(); err != nil {
				return nil, err
			}

			return params, nil
		}))
}
