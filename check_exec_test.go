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
		Path: "/bin/true",
	}

	err := c.Execute(1 * time.Second)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func TestCheckExecExecuteNonZero(t *testing.T) {
	c := CheckExec{Path: "/bin/false"}

	err := c.Execute(1 * time.Second)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckExecExecuteTimeout(t *testing.T) {
	c := CheckExec{
		Path: "/bin/sleep",
		Args: []string{"1"},
	}

	err := c.Execute(250 * time.Millisecond)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
