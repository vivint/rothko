// Copyright (C) 2017. See AUTHORS.

package outliers // import "github.com/spacemonkeygo/rothko/data/outliers"

import (
	"bytes"

	"github.com/spacemonkeygo/rothko/data"
)

func byteCopy(x []byte) []byte {
	return append([]byte(nil), x...)
}

func newOutlier(instance_id []byte, value float64) data.Outlier {
	return data.Outlier{
		InstanceId: byteCopy(instance_id),
		Value:      value,
	}
}

// InsertMin inserts the value for the instance id into the slice. If the
// instance id is present already, the value is updated. If the value is
// larger than all of the values present, it is appended to min up to the
// capacity of the slice.
func InsertMin(min []data.Outlier,
	instance_id []byte, value float64) []data.Outlier {

	// first search for the instance id. if it's in there, then overwrite it if
	// it's lower. if it's in there, we don't need to add it, so be done.
	for i, cand := range min {
		if !bytes.Equal(instance_id, cand.InstanceId) {
			continue
		}
		if value < cand.Value {
			min[i] = newOutlier(instance_id, value)
		}
		return min
	}

	if len(min) < cap(min) {
		min = append(min, newOutlier(instance_id, value))
	}

	for i, cand := range min {
		if value < cand.Value {
			copy(min[i+1:], min[i:])
			min[i] = newOutlier(instance_id, value)
			return min
		}
	}

	return min
}

// InsertMax inserts the value for the instance id into the slice. If the
// instance id is present already, the value is updated. If the value is
// smaller than all of the values present, it is appended to max up to the
// capacity of the slice.
func InsertMax(max []data.Outlier,
	instance_id []byte, value float64) []data.Outlier {

	// first search for the instance id. if it's in there, then overwrite it if
	// it's higher. if it's in there, we don't need to add it, so be done.
	for i, cand := range max {
		if !bytes.Equal(instance_id, cand.InstanceId) {
			continue
		}
		if value > cand.Value {
			max[i] = newOutlier(instance_id, value)
		}
		return max
	}

	if len(max) < cap(max) {
		max = append(max, newOutlier(instance_id, value))
	}

	for i, cand := range max {
		if value > cand.Value {
			copy(max[i+1:], max[i:])
			max[i] = newOutlier(instance_id, value)
			return max
		}
	}

	return max
}
