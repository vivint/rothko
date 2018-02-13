// Copyright (C) 2018. See AUTHORS.

package merge

import (
	"context"

	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/dist/tdigest"
)

// TODO(jeff): don't merge into tdigest, merge into a flat weighted buffer
// sampled from the distribution's cdf.

// MergeOptions are the arguments passed to Merge.
type MergeOptions struct {
	// Params are the parameters for the output distribution the merged record
	// should have.
	Params tdigest.Params

	// Records are the set of records to merge.
	Records []data.Record
}

// Merge combines the records into one large record. The seed is used to do
// deterministic merging.
func Merge(ctx context.Context, opts MergeOptions) (
	out data.Record, err error) {

	if len(opts.Records) == 0 {
		return out, Error.New("passed no records")
	}

	// merge the start and end time
	out.StartTime = opts.Records[0].StartTime
	out.EndTime = opts.Records[0].EndTime
	for _, r := range opts.Records[1:] {
		if r.StartTime < out.StartTime {
			out.StartTime = r.StartTime
		}
		if r.EndTime > out.EndTime {
			out.EndTime = r.EndTime
		}
	}

	// merge the observations
	for _, r := range opts.Records {
		out.Observations += r.Observations
	}

	// merge the distributions
	res := newResampler(opts.Params.NewUnwrapped())
	for _, r := range opts.Records {
		err := res.Sample(ctx, r)
		if err != nil {
			return out, err
		}
	}
	out.Distribution, out.Kind, err = res.Finish(ctx)
	if err != nil {
		return out, err
	}

	// merge the max and min values
	out.Min, out.Max = opts.Records[0].Min, opts.Records[0].Max
	for _, r := range opts.Records[1:] {
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
	for _, r := range opts.Records {
		// back compat: there may have been records without the merged field.
		// if merged is 0, that means it actually is 1.
		if r.Merged == 0 {
			r.Merged = 1
		}
		out.Merged += r.Merged
	}

	return out, nil
}
