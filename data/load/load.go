// Copyright (C) 2018. See AUTHORS.

package load

import (
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/dist"
	"github.com/spacemonkeygo/rothko/dist/tdigest"
	"github.com/zeebo/errs"
)

// TODO(jeff): look it up in the registrations.
// TODO(jeff): this package is in a weird spot.

// Load returns the dist.Dist for the data.Record.
func Load(rec data.Record) (dist.Dist, error) {
	switch rec.Kind {
	case "tdigest":
		return tdigest.Params{}.Unmarshal(rec.Distribution)
	}

	return nil, errs.New("unknown distribution kind: %q", rec.Kind)
}
