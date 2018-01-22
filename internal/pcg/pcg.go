// Copyright (C) 2018. See AUTHORS.

package pcg

type PCG struct {
	state uint64
	inc   uint64
}

// mul is the multiplier of the LCG step
const mul = 6364136223846793005

func New(state, inc uint64) PCG {
	// this code is equiv to initializing a PCG with a 0 state and the updated
	// inc and running
	//
	//    p.Uint32()
	//    p.state += state
	//    p.Uint32()
	//
	// to get the generator started

	inc = inc<<1 | 1
	return PCG{
		state: (inc+state)*mul + inc,
		inc:   inc,
	}
}

func (p *PCG) Uint32() uint32 {
	// this branch will be predicted to be false in most cases and so is
	// essentially free. this causes the zero value of a PCG to be the same as
	// New(0, 0).
	if p.inc == 0 {
		*p = New(0, 0)
	}

	// update the state (LCG step)
	oldstate := p.state
	p.state = oldstate*mul + p.inc

	// apply the output permutation to the old state
	xorshifted := uint32(((oldstate >> 18) ^ oldstate) >> 27)
	rot := uint32(oldstate >> 59)
	return xorshifted>>rot | (xorshifted << ((-rot) & 31))
}
