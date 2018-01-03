// Copyright (C) 2017. See AUTHORS.

package merge

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/merge/internal/randmarshal"
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
	switch r.DistributionKind {
	case data.DistributionKind_TDigest:
		other, err := tdigest.FromBytes(r.Distribution)
		if err != nil {
			return err
		}
		return res.dig.Merge(other)

	// more complicated. have to inspect the buffers and add them individually
	case data.DistributionKind_Random:
		var fr randmarshal.FinishedRandom
		err := proto.Unmarshal(r.Distribution, &fr)
		if err != nil {
			return Error.Wrap(err)
		}

		for _, buf := range fr.Buffers {
			if buf.Level > 31 {
				return Error.New("buffer has too much data")
			}
			count := uint32(1) << uint32(buf.Level)
			for _, value := range buf.Data {
				if err := res.dig.AddWeighted(value, count); err != nil {
					return Error.Wrap(err)
				}
			}
		}

		return nil
	}

	return Error.New("unknown distribution kind: %v", r.DistributionKind)
}

func (res *resampler) Finish(ctx context.Context) (
	[]byte, data.DistributionKind, error) {

	return res.dig.Marshal(nil), data.DistributionKind_TDigest, nil
}
