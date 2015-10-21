package buddha

import (
	"testing"
)

func TestCommandExecute(t *testing.T) {
	cmd := Command{
		Path: "echo",
		Args: []string{"hello"},
		Stdout: func(line string) {
			if line != "hello" {
				t.Fatal("expected hello, got", line)
			}
		},
	}

	err := cmd.Execute()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}
