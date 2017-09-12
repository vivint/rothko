// Copyright (C) 2017. See AUTHORS.

package merge

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/merge/internal/randmarshal"
	"github.com/zeebo/tdigest"
)

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
				if err := res.dig.Add(value, count); err != nil {
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
