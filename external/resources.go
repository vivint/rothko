// Copyright (C) 2018 Space Monkey, Inc.

package external

// Resources is a collection of all the external resources. It implements all
// of the methods of the fields but in a nil-safe way.
type Resources struct {
	Logger  Logger
	Monitor Monitor
}

// Infow calls Logger.Infow if Logger is not nil.
func (r Resources) Infow(msg string, keyvals ...interface{}) {
	if r.Logger != nil {
		r.Logger.Infow(msg, keyvals...)
	}
}

// Errorw calls Logger.Errorw if Logger is not nil.
func (r Resources) Errorw(msg string, keyvals ...interface{}) {
	if r.Logger != nil {
		r.Logger.Errorw(msg, keyvals...)
	}
}

// Observe calls Monitor.Observe if Logger is not nil.
func (r Resources) Observe(name string, value float64) {
	if r.Monitor != nil {
		r.Monitor.Observe(name, value)
	}
}
