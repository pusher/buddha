package flock

import (
	"os"
)

// Behaves like a *os.File but is only acquired with an exclusive flock().
//
// When the file is closed, the file is removed and the lock is released.
type FileLock struct {
	*os.File
}

// Either returns a flock-ed file or an error
//
// The lock is released by f.Close()-ing the file.
func Open(path string) (f *FileLock, err error) {
	var f2 *os.File
	if f2, err = os.Create(path); err == nil {
		if err = flock(f2); err != nil {
			f2.Close()
		} else {
			f = &FileLock{f2}
		}
	}
	return
}

func (f *FileLock) Close() (err error) {
	if err = f.File.Close(); err == nil {
		// Only remove the file if we actually held it before
		err = os.Remove(f.File.Name())
	}
	return
}
