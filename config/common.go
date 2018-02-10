// Copyright (C) 2018. See AUTHORS.

package config

import (
	"time"

	"github.com/zeebo/errs"
)

// ParseError wraps all of the errors from parsing.
var ParseError = errs.Class("parse error")

// textDuration wraps a time.Duration providing the TextUnmarshaler interface.
type textDuration struct {
	time.Duration
}

func (d *textDuration) UnmarshalText(text []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
