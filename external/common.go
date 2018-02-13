// Copyright (C) 2018 Space Monkey, Inc.

package external

// Default is the default set of resources. Can be overridden by plugins.
var Default Resources

// Logger is used when logging is required. It is built to match the uber/zap
// SugaredLogger type.
type Logger interface {
	Infow(msg string, keyvals ...interface{})
	Errorw(msg string, keyvals ...interface{})
}

// Monitor is used to monitor rothko's operation.
type Monitor interface {
	Observe(name string, value float64)
}

//
// package level implementations
//

// Infow calls Infow on the default resources.
func Infow(msg string, keyvals ...interface{}) {
	Default.Infow(msg, keyvals...)
}

// Errorw calls Errorw on the default Resources.
func Errorw(msg string, keyvals ...interface{}) {
	Default.Errorw(msg, keyvals...)
}

// Observe calls Observe on the default Resources.
func Observe(name string, value float64) {
	Default.Observe(name, value)
}
