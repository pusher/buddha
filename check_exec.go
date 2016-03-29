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

func (c CheckExec) Execute(timeout time.Duration) (bool, error) {
	path, err := exec.LookPath(c.Path)
	if err != nil {
		return false, err
	}

	fullArgs := []string{c.Path}
	fullArgs = append(fullArgs, c.Args...)

	p, err := os.StartProcess(path, fullArgs, &os.ProcAttr{})
	if err != nil {
		return false, err
	}

	// assume the exit codes of the process mean 0 => true, 1 => false, 2 => error
	done := make(chan bool, 1)
	fail := make(chan error, 1)
	go func() {
		processState, err := p.Wait()
		if err != nil {
			fail <- err
		} else if processState.Success() {
			done <- true
		} else {
			if status, ok := processState.Sys().(syscall.WaitStatus); ok {
				switch status.ExitStatus() {
				case 1:
					done <- false
				case 2:
					fail <- fmt.Errorf("failed with exit code: 2")
				default:
					fail <- fmt.Errorf("unexpected exit code: %d", status.ExitStatus())
				}
			} else {
				fail <- fmt.Errorf("platform does not support exit codes")
			}
		}
	}()

	select {
	case result := <-done:
		return result, nil

	case err := <-fail:
		return false, err

	case <-time.After(timeout):
		p.Kill()
		return false, fmt.Errorf("timeout exceeded")
	}
}

func (c CheckExec) String() string {
	return c.Name
}
