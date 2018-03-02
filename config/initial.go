// Copyright (C) 2018. See AUTHORS.

package config

import "strings"

var InitialConfig = strings.TrimSpace(`
#
# Main configurations control overall behavior of rothko:
#
#	duration: how often the aggregated distributions are flushed to the
#	          database.
#
#	plugins: these files will be loaded at process start and can be used to
#	         add new kinds of databases or listeners. See the top rothko
#	         package documentation for how to create a plugin to add more kinds
#	         of listeners or databases.
#

[main]
	duration = "10m"
	plugins = [
		# "my_plugin.so",
	]

#
# Multiple listeners can be specified to receive data. There may be multiple
# kinds of listeners supported, but currently only the graphite wire protocol
# is built in.
#

[[listeners.graphite]]
	address = ":1111"

#
# example to add a second graphite listener:
#

# [[listeners.graphite]]
# 	address = ":2222"

#
# The files database keeps track of the metric data as a set of files. Each
# metric is allowed to have a certain number of files storing the data and
# each file is composed of a sequence of records, with a metadata record at
# the start. The size of the records, and the number of records per file is
# controlled by the size and cap fields. The maximum size of the database is
# thus (number of metrics) * size * (cap + 1) * (files + 1). The reason for
# files + 1 is to guarantee that we can create new files when deleting old
# ones. For every metric that receives a value during the duration, one record 
# will be written to the files that back it's data. Thus, the maximum amount of
# time metric data will be retained is given by cap * files * duration, 
# assuming that size is sufficient to hold a single record. Metric data will be
# split into multiple records, if necessary.
#

[database.files]
	directory = "data"
	size = 256
	cap = 400
	files = 2

#
# The files database allows some tuning:
#
#	buffer: controls how many records will be buffered while writing them to
#	        disk.
#
#	drop: if true, the process flushing the records to be written will drop
#	      records if they can not be immediately added to the buffer. otherwise
#	      it will block until the record is collected to be written.
#
#	workers: specifies the number of workers writing records to disk. If 0 or
#	         not set, will use GOMAXPROCS - 1. The number of workers should be
#	         less than GOMAXPROCS because they used mmap'd I/O to do writing,
#	         and the Go runtime cannot schedule around those memory accesses
#	         blocking.
#
#	handles: specifies the number of handles to keep in a cache for the metric
#	         files. If 0 or unspecified, then 1024 less than the soft limit of
#	         file handles as reported by getrlimit is used.
#

# [database.files.tuning]
# 	buffer = 20000
# 	drop = false
# 	workers = 0
# 	handles = 0

#
# The distribution sketch that the metrics will be stored with. A T-Digest
# implementation is provided, but more can be added with plugins.
#

[dist.tdigest]
	compression = 5.0

#
# The server runs an API for querying the metrics, as well as a web interface
# for rendering and interacting. The address is the port that the server will
# listen on, and the origin is used to handle CORS. You may want to limit it
# in a production deploy.
#

[api]
	address = ":8080"
	origin = "*"

#
# If the api.tls section is specified, it will listen with TLS using the
# provided key and cert.
#

# [api.tls]
# 	key = "key"
# 	cert = "cert"

#
# If the api.security section is specified, the resources will all be protected
# by http basic auth. Consider using the api.tls section if you use this as
# http basic auth sends the credentials in the clear.
#

# [api.security]
# 	username = "admin"
# 	password = "boogers"
`)
