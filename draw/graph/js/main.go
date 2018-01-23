// Copyright (C) 2018. See AUTHORS.

package main

import (
	"context"

	"github.com/gopherjs/gopherjs/js"
	"github.com/zeebo/errs"
)

var (
	ctx      = context.Background()
	panicErr = errs.Class("panic")
)

func main() {
	js.Global.Get("self").Call("addEventListener", "message", render)
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

type (
	D = map[string]interface{}
	L = []interface{}
)
