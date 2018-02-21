// Copyright (C) 2018. See AUTHORS.

package graphite

import (
	"context"

	"github.com/vivint/rothko/internal/typeassert"
	"github.com/vivint/rothko/listener"
	"github.com/vivint/rothko/registry"
)

func init() {
	registry.RegisterListener("graphite", registry.ListenerMakerFunc(
		func(ctx context.Context, config interface{}) (listener.Listener, error) {
			a := typeassert.A(config)
			lis := New(a.I("address").String())
			if err := a.Err(); err != nil {
				return nil, err
			}

			return lis, nil
		}))
}
