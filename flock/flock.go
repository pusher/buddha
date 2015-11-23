package flock

import (
	"errors"
	"os"
)

var (
	ErrLocked = errors.New("path already locked")
)

type Flock string

// acquire file based lock constraint, or return ErrLocked
func Lock(path string) (Flock, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return Flock(path), ErrLocked
		}

		return Flock(path), err
	}
	defer file.Close()

	return Flock(path), nil
}

// remove file based lock constraint
func (f Flock) Close() error {
	return os.Remove(string(f))
}
