// Copyright (C) 2018. See AUTHORS.

package load

import (
	"context"

	"github.com/vivint/rothko/data"
	"github.com/vivint/rothko/dist"
	"github.com/vivint/rothko/registry"
)

// TODO(jeff): this package is in a weird spot.

// Load returns the dist.Dist for the data.Record.
func Load(ctx context.Context, rec data.Record) (dist.Dist, error) {
	params, err := registry.NewDistribution(ctx, rec.Kind, nil)
	if err != nil {
		return nil, err
	}
	return params.Unmarshal(rec.Distribution)
}
