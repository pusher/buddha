package buddha

import (
	"fmt"
	"time"
	"os"
	"os/exec"
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
	path, err := exec.LookPath(c.Path)
	if err != nil {
		return err
	}

	p, err := os.StartProcess(path, c.Args, &os.ProcAttr{})
	if err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		_, err := p.Wait()
		done <- err
	}()

	select {
	case err := <-done:
		return err

	case <-time.After(timeout):
		p.Kill()
		return fmt.Errorf("timeout exceeded")
	}
}

func (c CheckExec) String() string {
	return c.Name
}
