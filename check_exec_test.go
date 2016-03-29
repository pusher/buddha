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

	result, err := c.Execute(1 * time.Second)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if result != true {
		t.Fatal("expected result to be true, got false")
	}
}

func TestCheckExecExecuteReturn1(t *testing.T) {
	c := CheckExec{Path: "false"}

	result, err := c.Execute(1 * time.Second)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if result != false {
		t.Fatal("expected result to be false, got true")
	}
}

func TestCheckExecExecuteReturn2(t *testing.T) {
	c := CheckExec{Path: "sh -c 'exit 2'"}

	_, err := c.Execute(1 * time.Second)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckExecExecuteTimeout(t *testing.T) {
	c := CheckExec{
		Path: "/bin/sleep",
		Args: []string{"1"},
	}

	_, err := c.Execute(250 * time.Millisecond)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
