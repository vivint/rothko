// Copyright (C) 2018. See AUTHORS.

package junk

import (
	"sync"

	"github.com/zeebo/errs"
)

type Flag struct {
	mu sync.Mutex
	on bool
}

func (f *Flag) Start() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.on {
		return errs.New("already started")
	}
	f.on = true
	return nil
}

func (f *Flag) Stop() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.on = false
}
