// Copyright (C) 2018. See AUTHORS.

package data

import "github.com/vivint/rothko/dist"

type fakeParams struct{ dist.Params }

func (f fakeParams) New() (dist.Dist, error) { return fakeDist{}, nil }
func (f fakeParams) Kind() string            { return "fake" }

type fakeDist struct{ dist.Dist }

func (f fakeDist) Kind() string               { return "fake" }
func (f fakeDist) Observe(float64)            {}
func (f fakeDist) Marshal(data []byte) []byte { return append(data, 0) }
