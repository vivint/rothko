// Copyright (C) 2017. See AUTHORS.

// package mmap provides a lower level version of mmap than syscall.
package mmap // import "github.com/spacemonkeygo/rothko/disk/files/internal/mmap"

import (
	"github.com/spacemonkeygo/errors"
)

var (
	Error = errors.NewClass("mmap")
)
