// Copyright (C) 2017. See AUTHORS.

package merge

import (
	"context"

	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/zeebo/tdigest"
	monkit "gopkg.in/spacemonkeygo/monkit.v2"
)

var (
	mon = monkit.Package()

	Error = errors.NewClass("merge")
)

// MergeCompression is the compression value for a tdigest that will be the
// output distribution for any merge operation.
const MergeCompression = 10

// Merge combines the records into one large record. The seed is used to do
// deterministic merging.
func Merge(ctx context.Context, seed uint64, rs ...data.Record) (
	out data.Record, err error) {
	defer mon.Task()(&ctx)(&err)

	if len(rs) == 0 {
		return out, Error.New("passed no records")
	}

	// merge the start and end time
	out.StartTime, out.EndTime = rs[0].StartTime, rs[0].EndTime
	for _, r := range rs[1:] {
		if r.StartTime < out.StartTime {
			out.StartTime = r.StartTime
		}
		if r.EndTime > out.EndTime {
			out.EndTime = r.EndTime
		}
	}

	// merge the observations
	for _, r := range rs {
		out.Observations += r.Observations
	}

	// merge the distributions
	res := newResampler(tdigest.New(MergeCompression))
	for _, r := range rs {
		err := res.Sample(ctx, r)
		if err != nil {
			return out, err
		}
	}
	out.Distribution, out.DistributionKind, err = res.Finish(ctx)
	if err != nil {
		return out, err
	}

	// merge the max and min values
	out.Min, out.Max = rs[0].Min, rs[0].Max
	for _, r := range rs[1:] {
		if r.Min < out.Min {
			out.Min = r.Min
			out.MinId = r.MinId
		}
		if r.Max > out.Max {
			out.Max = r.Max
			out.MaxId = r.MaxId
		}
	}

	// merge how many we've merged
	for _, r := range rs {
		// back compat: there may have been records without the merged field.
		// if merged is 0, that means it actually is 1.
		if r.Merged == 0 {
			r.Merged = 1
		}
		out.Merged += r.Merged
	}

	return out, nil
}
