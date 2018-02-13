// Copyright (C) 2018. See AUTHORS.

package typeassert

import (
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestAsserter(t *testing.T) {
	type (
		D = map[string]interface{}
		L = []interface{}
	)

	data := D{
		"int":    2,
		"bool":   true,
		"string": "foo",
		"list":   L{2, true, "foo"},
		"map":    D{"int": 2},
	}

	t.Run("Success", func(t *testing.T) {
		a := A(data)
		assert.Equal(t, a.I("int").Int(), 2)
		assert.Equal(t, a.I("bool").Bool(), true)
		assert.Equal(t, a.I("string").String(), "foo")
		assert.Equal(t, a.I("list").N(0).Int(), 2)
		assert.Equal(t, a.I("list").N(1).Bool(), true)
		assert.Equal(t, a.I("list").N(2).String(), "foo")
		assert.Equal(t, a.I("map").I("int").Int(), 2)
		assert.NoError(t, a.Err())
	})

	t.Run("Failure", func(t *testing.T) {
		{
			a := A(data)
			a.String()
			assert.Error(t, a.Err())
		}

		{
			a := A(data)
			a.I("int").String()
			assert.Error(t, a.Err())
		}

		{
			a := A(data)
			a.I("string").Int()
			assert.Error(t, a.Err())
		}

		{
			a := A(data)
			a.I("list").N(0).String()
			assert.Error(t, a.Err())
		}

		{
			a := A(data)
			a.I("map").I("int").String()
			assert.Error(t, a.Err())
		}

	})
}
