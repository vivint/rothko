// Copyright (C) 2017. See AUTHORS.

// package files implements a disk.Source and disk.Writer
package files // import "github.com/spacemonkeygo/rothko/disk/files"

import (
	"github.com/spacemonkeygo/errors"
	monkit "gopkg.in/spacemonkeygo/monkit.v2"
)

var (
	Error = errors.NewClass("files")

	mon = monkit.Package()
)
