// +build linux plan9 freebsd solaris

package lsof

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/pkg/errors"
)

// Scan scans the global /proc/:pid/fd and returns PIDs that are reading files.
// The files are checked using the given StringChecker callback, which must be
// thread-safe.
func Scan(chkFn StringChecker) (map[string][]int, error) {
	pidList, err := quickListDir("/proc")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read /proc/")
	}

	var results = map[string][]int{}

	if NumWorkers < 2 || len(pidList) < ChunkSize {
		tryReadFdDir(pidList, chkFn, results, nil)

	} else {
		// Try and distribute the load across workers.
		var mutex sync.Mutex
		var job = make(chan []string, NumWorkers) // channel of pidLists

		group := sync.WaitGroup{}
		group.Add(NumWorkers)

		for i := 0; i < NumWorkers; i++ {
			go func() {
				for pidList := range job {
					tryReadFdDir(pidList, chkFn, results, &mutex)
				}
				group.Done()
			}()
		}

		var end = ChunkSize

		for i := 0; i < len(pidList); {
			job <- pidList[i:end]

			i = end

			if end += ChunkSize; end > len(pidList) {
				end = len(pidList)
			}
		}

		close(job)
		group.Wait()
	}

	return results, nil
}

func tryReadFdDir(pidList []string, fn StringChecker, res map[string][]int, mu *sync.Mutex) {
	for _, pid := range pidList {
		p, err := strconv.Atoi(pid)
		if err != nil {
			continue
		}

		readFdDir("/proc/"+pid+"/fd", p, fn, res, mu)
	}
}

func readFdDir(fdDir string, pid int, fn StringChecker, res map[string][]int, mu *sync.Mutex) {
	symList, err := quickListDir(fdDir)
	if err != nil {
		return
	}

	for _, fd := range symList {
		p, err := os.Readlink(filepath.Join(fdDir, fd))
		if err != nil {
			continue
		}

		if !fn(p) {
			continue
		}

		if mu != nil {
			mu.Lock()
			res[p] = append(res[p], pid)
			mu.Unlock()
		} else {
			res[p] = append(res[p], pid)
		}
	}

	return
}

func quickListDir(dirPath string) ([]string, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open dir %q", dirPath)
	}
	defer f.Close()

	return f.Readdirnames(-1)
}
