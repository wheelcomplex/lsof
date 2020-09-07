package lsof

import (
	"errors"
	"runtime"
)

// StringChecker is used to check if a resolved path matches. The callback MUST
// be thread-safe.
type StringChecker func(path string) bool

// ChunkSize is the number of files to scan sequentially per worker. It is also
// the threshold before workers are utilized.
//
// 5 has been found to be the sweet spot.
const ChunkSize = 5

// NumWorkers is the number of workers to distribute scanning to.
var NumWorkers = runtime.GOMAXPROCS(-1)

var ErrUnsupportedOS = errors.New("unsupported OS")
