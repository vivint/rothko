// Copyright (C) 2017. See AUTHORS.

package merge

import (
	"math/rand"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/merge/internal/randmarshal"
	"github.com/spacemonkeygo/rothko/internal/assert"
	"github.com/zeebo/tdigest"
	random "gopkg.in/spacemonkeygo/random.v1"
)

func TestResampler(t *testing.T) {
	const count = 1000000

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	rn := random.NewRandom(0.01)
	for i := 0; i < count; i++ {
		rn.Add(rng.NormFloat64())
	}
	sum := rn.Summarize()
	buf, err := marshal(rn)
	assert.NoError(t, err)

	res := newResampler(tdigest.New(10))
	assert.NoError(t, res.Sample(ctx, data.Record{
		Distribution:     buf,
		DistributionKind: data.DistributionKind_Random,
	}))

	// i don't really want to do the statistical tests to validate this, so
	// just log it and inspect that it's approximately close.
	for q := 0.0; q <= 1.0; q += 1.0 / 128 {
		t.Logf("%9.6f %9.6f %9.6f", q, res.dig.Quantile(q), sum.Query(q))
	}
}

//
// conversion functions for marshaling
//

func convertFinishedRandom(fr random.FinishedRandom) (
	frm randmarshal.FinishedRandom) {

	return randmarshal.FinishedRandom{
		E:       fr.E,
		N:       fr.N,
		Buffers: convertBuffers(fr.Buffers),
	}
}

func convertBuffers(bs []random.Buffer) []randmarshal.Buffer {
	out := make([]randmarshal.Buffer, len(bs))
	for i, b := range bs {
		out[i] = randmarshal.Buffer(b)
	}
	return out
}

func marshal(rn *random.Random) ([]byte, error) {
	frv := convertFinishedRandom(rn.Finish())
	return proto.Marshal(&frv)
}
