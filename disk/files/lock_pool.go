// Copyright (C) 2017. See AUTHORS.

package files

import "sync"

// TODO(jeff): we may want some caching around lockPoolState to avoid
// allocations on frequently accessed keys.

type lockPoolState struct {
	mu  sync.Mutex
	ref int
}

type lockPool struct {
	mu    sync.Mutex
	locks map[string]*lockPoolState
}

func newLockPool() *lockPool {
	return &lockPool{
		locks: make(map[string]*lockPoolState),
	}
}

func (l *lockPool) Lock(key string) {
	l.mu.Lock()
	st, ok := l.locks[key]
	if !ok {
		st = new(lockPoolState)
		l.locks[key] = st
	}
	st.ref++
	l.mu.Unlock()

	st.mu.Lock()
}

func (l *lockPool) Unlock(key string) {
	l.mu.Lock()
	st, ok := l.locks[key]
	if !ok {
		panic("unlock with no state")
	}
	st.ref--
	if st.ref == 0 {
		delete(l.locks, key)
	}
	l.mu.Unlock()

	st.mu.Unlock()
}
