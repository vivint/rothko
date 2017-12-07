// Copyright (C) 2017. See AUTHORS.

package accept

import (
	"context"

	"github.com/spacemonkeygo/rothko/data/scribble"
)

// Acceptrix is a type that reads from some data source and pushes the data
// into the scribbler.
type Acceptrix interface {
	// Run should scribble the data into the provided Scribbler until the
	// context is canceled.
	Run(ctx context.Context, scr *scribble.Scribbler) error
}

// AcceptrixMaker returns a new Acceptrix for the given config string.
type AcceptrixMaker func(ctx context.Context, config string) (Acceptrix, error)
