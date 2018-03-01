// Copyright (C) 2018. See AUTHORS.

// package ui provides a gzipped tar archive of the compiled ui.
package ui

// Tarball contains a gzipped tar archive to be served for the ui. If it is
// nil, no ui is served. If you want it to not be nil, use `roth generate`.
var Tarball []byte
