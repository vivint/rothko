// Copyright (C) 2017. See AUTHORS.

package draw

import (
	"context"
	"math"
	"math/rand"
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

var ctx = context.Background()

func TestFastFloor(t *testing.T) {
	// this test passes for higher values, but we should never have more than
	// 1000 colors.
	for i := 0; i < 1000; i++ {
		for j := 0; j < 100; j++ {
			f := rand.Float64()
			v := float64(i) + f
			assert.Equal(t, int(math.Floor(v)), fastFloor(v))
		}
	}
}
