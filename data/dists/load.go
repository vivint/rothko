// Copyright (C) 2017. See AUTHORS.

package dists

import (
	"github.com/spacemonkeygo/rothko/data"
	tdigest_wrapper "github.com/spacemonkeygo/rothko/data/dists/tdigest"
	"github.com/zeebo/tdigest"
)

// Load returns the Dist for the record.
func Load(rec data.Record) (Dist, error) {
	switch rec.DistributionKind {
	case data.DistributionKind_TDigest:
		other, err := tdigest.FromBytes(rec.Distribution)
		if err != nil {
			return nil, err
		}
		return tdigest_wrapper.Wrap(other), nil

	// more complicated. punt for now.
	case data.DistributionKind_Random:
		return nil, Error.New("TODO: implement random querying")
	}

	return nil, Error.New("unknown distribution kind: %v",
		rec.DistributionKind)
}
