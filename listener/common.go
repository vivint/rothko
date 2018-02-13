// Copyright (C) 2018. See AUTHORS.

package listener

import (
	"context"

	"github.com/spacemonkeygo/rothko/data"
)

// Listener is a type that writes from some data source to the privided Writer.
type Listener interface {
	// Run should Add values into the Writer until the context is canceled.
	Run(ctx context.Context, w *data.Writer) (err error)
}
