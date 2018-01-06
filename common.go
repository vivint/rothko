// Copyright (C) 2017. See AUTHORS.

package rothko

import (
	"context"

	"github.com/spacemonkeygo/rothko/data/scribble"
)

// Logger is used when logging is required.
type Logger interface {
	Log(keyvals ...interface{}) error
}

// Monitor is used to monitor rothko's operation.
type Monitor interface {
	Task(name string) func(*error)
}

// Acceptrix is a type that reads from some data source and pushes the data
// into the scribbler.
type Acceptrix interface {
	// Run should scribble the data into the provided Scribbler until the
	// context is canceled.
	Run(ctx context.Context, scr *scribble.Scribbler) error
}
