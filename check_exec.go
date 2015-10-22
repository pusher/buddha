package buddha

import (
	"fmt"
	"os/exec"
	"time"
)

type CheckExec struct {
	// name of check in logs
	Name string `json:"name"`

	// path to executable
	Path string `json:"path"`

	// arguments to pass to executable
	Args []string `json:"args"`
}

func (c CheckExec) Validate() error {
	if len(c.Path) == 0 {
		return fmt.Errorf("expected command to execute")
	}

	return nil
}

func (c CheckExec) Execute(timeout time.Duration) error {
	cmd := exec.Command(c.Path, c.Args...)

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			return err
		}

	case <-time.After(timeout):
		return fmt.Errorf("timeout exceeded")
	}

	return nil
}

func (c CheckExec) String() string {
	return c.Name
}
