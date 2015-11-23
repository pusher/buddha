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

func (j Jobs) Len() int           { return len(j) }
func (j Jobs) Swap(i, n int)      { j[i], j[n] = j[n], j[i] }
func (j Jobs) Less(i, n int) bool { return j[i].Name < j[n].Name }

// return new array of jobs matching name filter
func (j Jobs) Filter(f []string) Jobs {
	var n Jobs
	for i := 0; i < len(j); i++ {
		if inArray(f, j[i].Name) {
			n = append(n, j[i])
		}
	}

	return n
}

type Job struct {
	// name of job in logs
	Name string `json:"name"`

	// true if root privileges are required to run
	Root bool `json:"root"`

	// commands to execute
	Commands []Command `json:"commands"`
}

// open job config from reader
func Open(r io.Reader) (Jobs, error) {
	var jobs Jobs
	err := json.NewDecoder(r).Decode(&jobs)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

// open job config from file
func OpenFile(filename string) (Jobs, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Open(file)
}

// open job config files from directory
func OpenDir(dirname string) (Jobs, error) {
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

	return jobs, nil
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

// return true if element s in array a
func inArray(a []string, s string) bool {
	for i := 0; i < len(a); i++ {
		if a[i] == s {
			return true
		}
	}

	return false
}
