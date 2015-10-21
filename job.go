package buddha

import (
	"encoding/json"
	"io"
	"os"
)

type Jobs []*Job

type Job struct {
	// name of job in logs
	Name string `json:"name"`

	// commands to execute
	Commands []Command `json:"commands"`
}

// open job config from reader
func Open(r io.Reader) (*Jobs, error) {
	var jobs Jobs
	err := json.NewDecoder(r).Decode(&jobs)
	if err != nil {
		return nil, err
	}

	return &jobs, nil
}

// open job config from file
func OpenFile(filename string) (*Jobs, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Open(file)
}
