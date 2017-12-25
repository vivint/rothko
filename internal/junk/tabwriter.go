// Copyright (C) 2017. See AUTHORS.

package junk // import "github.com/spacemonkeygo/rothko/internal/junk"

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/zeebo/errs"
)

type Tabbed struct {
	tw  *tabwriter.Writer
	err error
}

func NewTabbed(w io.Writer) *Tabbed {
	return &Tabbed{
		tw: tabwriter.NewWriter(w, 0, 8, 3, ' ', 0),
	}
}

func (t *Tabbed) Write(values ...string) {
	if t.err == nil {
		_, t.err = fmt.Fprintln(t.tw, strings.Join(values, "\t"))
		t.err = errs.Wrap(t.err)
	}
}

func (t *Tabbed) Flush() {
	if t.err == nil {
		t.err = errs.Wrap(t.tw.Flush())
	}
}

func (t *Tabbed) Error() error {
	return t.err
}
