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
		Path: "true", // The `true` command
	}

	err := c.Execute(1 * time.Second)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func TestCheckExecExecuteReturn1(t *testing.T) {
	c := CheckExec{Path: "false"}

	err := c.Execute(1 * time.Second)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(CheckFalse); !ok {
		t.Fatal("expected err to be CheckFalse")
	}
}

func TestCheckExecExecuteReturn2(t *testing.T) {
	c := CheckExec{Path: "sh -c 'exit 2'"}

	err := c.Execute(1 * time.Second)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(CheckFalse); ok {
		t.Fatal("expected err to not be CheckFalse")
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
	if _, ok := err.(CheckFalse); ok {
		t.Fatal("expected err to not be CheckFalse")
	}
}
