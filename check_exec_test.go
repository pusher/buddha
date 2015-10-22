package buddha

import (
	"testing"
	"time"
)

func TestCheckExecValidate(t *testing.T) {
	c1 := CheckExec{}
	if err := c1.Validate(); err == nil {
		t.Fatal("expected error, got nil")
	}

	c2 := CheckExec{Path: "foobar"}
	if err := c2.Validate(); err != nil {
		t.Fatal("expected nil, got", err)
	}
}

func TestCheckExecExecute(t *testing.T) {
	c := CheckExec{
		Path: "echo",
		Args: []string{"hello"},
	}

	err := c.Execute(1 * time.Second)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}
