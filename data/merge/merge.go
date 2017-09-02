// Copyright (C) 2017. See AUTHORS.

package merge //  import "github.com/spacemonkeygo/rothko/data/merge"

import (
	"context"

	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/outliers"
	monkit "gopkg.in/spacemonkeygo/monkit.v2"
)

var (
	mon = monkit.Package()

	Error = errors.NewClass("merge")
)

// Merge combines the records into one large record. The seed is used to do
// deterministic merging.
func Merge(ctx context.Context, seed uint64, rs ...data.Record) (
	out data.Record, err error) {
	defer mon.Task()(&ctx)(&err)

	if len(rs) == 0 {
		return out, Error.New("passed no records")
	}

	// TODO(jeff): merge the distributions. do something if they are different
	// kinds, like resample? there's probably something we can do there with
	// the t-digest, since it allows weighted additions. what could the merge
	// api even be then? hmm...

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

	// merge the max and min values
	out.Min, out.Max = rs[0].Min, rs[0].Max
	for _, r := range rs[1:] {
		if r.Min < out.Min {
			out.Min = r.Min
		}
		if r.Max > out.Max {
			out.Max = r.Max
		}
	}

	// merge the outliers
	for _, r := range rs {
		if cand := r.Outliers; cand > r.Outliers {
			r.Outliers = cand
		}
	}

	out.Smallest = make([]data.Outlier, 0, out.Outliers)
	out.Largest = make([]data.Outlier, 0, out.Outliers)

	for _, r := range rs {
		out.Smallest, out.Largest = mergeOutliers(r, out.Smallest, out.Largest)
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

// mergeOutliers takes outliers from the given record and merges them into
// the smallest and largest collection of outliers, returning them.
func mergeOutliers(r data.Record, smallest, largest []data.Outlier) (
	[]data.Outlier, []data.Outlier) {

	for _, out := range r.Smallest {
		smallest = outliers.InsertMin(smallest, out.InstanceId, out.Value)
	}
	for _, out := range r.Largest {
		largest = outliers.InsertMax(largest, out.InstanceId, out.Value)
	}
	return smallest, largest
}
