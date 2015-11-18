package buddha

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

// open job config files from directory
func OpenDir(dirname string) (*Jobs, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	var jobs Jobs
	for _, file := range files {
		path := filepath.Join(dirname, file.Name())

		if file.IsDir() || !strings.HasSuffix(path, ".json") {
			continue
		}

		var njobs Jobs
		err := unmarshalFile(path, &njobs)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, njobs...)
	}

	return &jobs, nil
}

// unmarshal json file to interface
func unmarshalFile(filename string, v interface{}) error {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(v)
}
