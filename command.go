package buddha

import (
	"bufio"
	"io"
	"os/exec"
)

type Command struct {
	// path to executable
	Path string `json:"path"`

	// arguments to pass to executable
	Args []string `json:"args,omitempty"`

	Before Checks `json:"before"`
	After  Checks `json:"after"`

	// timeout between executing command and beginning health checking
	Grace Duration `json:"grace"`

	// maximum time for check execution
	Timeout Duration `json:"timeout"`

	// timeout between health checks
	Interval Duration `json:"interval"`

	// maximum health check failures before failing run
	Failures int `json:"failures"`

	Stdout func(line string) `json:"-"` // call func for each stdout line
}

// execute system command, piping logs to reader
func (c Command) Execute() error {
	cmd := exec.Command(c.Path, c.Args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer stderr.Close()

	// line readers for log data
	if c.Stdout != nil {
		go c.lineReader(stdout)
		go c.lineReader(stderr)
	}

	return cmd.Run()
}

// execute stdout function for each line of output
func (c Command) lineReader(r io.Reader) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		c.Stdout(scanner.Text())
	}
}
