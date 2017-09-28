// Copyright (C) 2017. See AUTHORS.

package pcg

type PCG struct {
	state uint64
	inc   uint64
}

func New(state, inc uint64) PCG {
	p := PCG{
		state: 0,
		inc:   inc<<1 | 1,
	}

	p.Uint32()
	p.state += state
	p.Uint32()

	return p
}

func (p *PCG) Uint32() uint32 {
	// we unroll the seeding steps in New above, careful to avoid recursive
	// calls so that the method may be inlined. this branch will be predicted
	// to be false in most cases and so is essentially free. this causes the
	// zero value of a PCG to be the same as New(0, 0).
	if p.inc == 0 {
		p.inc = 1
		p.state = p.state*6364136223846793005 + 1
		p.state = p.state*6364136223846793005 + 1
	}

	// update the state (LCG step)
	oldstate := p.state
	p.state = oldstate*6364136223846793005 + p.inc

	// apply the output permutation to the old state
	xorshifted := uint32(((oldstate >> 18) ^ oldstate) >> 27)
	rot := uint32(oldstate >> 59)
	return xorshifted>>rot | (xorshifted << ((-rot) & 31))
}
