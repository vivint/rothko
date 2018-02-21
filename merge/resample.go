// Copyright (C) 2018. See AUTHORS.

package merge

import (
	"context"

	"github.com/vivint/rothko/data"
	"github.com/zeebo/tdigest"
)

//
// a resampler is used to merge a bunch of unknown distribution kinds into a
// single t-digest, since t-digest seems to be the best so far. hopefully
// it will always be the case that we can support this operation with whatever
// distributions we have in the future.
//

// TODO(jeff): this can probably use some refactoring to use the Dist
// abstraction and data/dists/... packages. At the same time, these details
// are going to be specific to the concrete distributions, so perhaps it's ok.

type resampler struct {
	dig *tdigest.TDigest
}

func newResampler(dig *tdigest.TDigest) *resampler {
	return &resampler{
		dig: dig,
	}
}

func (res *resampler) Sample(ctx context.Context, r data.Record) error {
	switch r.Kind {
	case "tdigest":
		other, err := tdigest.FromBytes(r.Distribution)
		if err != nil {
			return err
		}
		return res.dig.Merge(other)
	}

	return Error.New("unknown distribution kind: %v", r.Kind)
}

func (res *resampler) Finish(ctx context.Context) ([]byte, string, error) {
	return res.dig.Marshal(nil), "tdigest", nil
}
