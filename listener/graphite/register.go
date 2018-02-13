// Copyright (C) 2018. See AUTHORS.

package graphite

import (
	"context"

	"github.com/spacemonkeygo/rothko/internal/typeassert"
	"github.com/spacemonkeygo/rothko/listener"
	"github.com/spacemonkeygo/rothko/registry"
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
