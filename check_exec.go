package buddha

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
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
	path, err := exec.LookPath(c.Path)
	if err != nil {
		return err
	}

	fullArgs := []string{c.Path}
	fullArgs = append(fullArgs, c.Args...)

	p, err := os.StartProcess(path, fullArgs, &os.ProcAttr{})
	if err != nil {
		return err
	}

	// assume the exit codes of the process mean 0 => true, 1 => false, 2 => error
	fail := make(chan error, 1)
	go func() {
		processState, err := p.Wait()
		if err != nil {
			fail <- err
		} else if processState.Success() {
			fail <- nil
		} else {
			if status, ok := processState.Sys().(syscall.WaitStatus); ok {
				switch status.ExitStatus() {
				case 1:
					// The check failed in an expected way
					fail <- CheckFalse(fmt.Sprintf("command `%s` returned exit code 1", path))
				case 2:
					// The command had an unexpected error
					fail <- fmt.Errorf("command `%s` failed with exit code: 2", path)
				default:
					fail <- fmt.Errorf("command `%s` returned unexpected exit code: %d", path, status.ExitStatus())
				}
			} else {
				fail <- fmt.Errorf("platform does not support exit codes")
			}
		}
	}()

	select {
	case err := <-fail:
		return err

	case <-time.After(timeout):
		p.Kill()
		return fmt.Errorf("timeout exceeded")
	}
}

func (c CheckExec) String() string {
	return c.Name
}
