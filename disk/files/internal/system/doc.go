// Copyright (C) 2017. See AUTHORS.

// package system provides optimized and dangerous functions for system calls.
package system // import "github.com/spacemonkeygo/rothko/disk/files/internal/system"

import (
	"github.com/spacemonkeygo/errors"
)

var (
	Error = errors.NewClass("system")
)
