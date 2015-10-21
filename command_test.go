package buddha

import (
	"testing"
)

func TestCommandAll(t *testing.T) {
	c := Command{
		HTTP: []CheckHTTP{
			CheckHTTP{Name: "foo"},
		},

		TCP: []CheckTCP{
			CheckTCP{Name: "bar"},
		},
	}

	checks := c.All()
	if len(checks) != 2 {
		t.Fatal("expected 2 checks, got", len(checks))
	}

	if http, ok := checks[0].(CheckHTTP); ok {
		if http.Name != "foo" {
			t.Fatal("expected http check foo, got", http.Name)
		}
	} else {
		t.Fatal("expected checks[0] http")
	}

	if tcp, ok := checks[1].(CheckTCP); ok {
		if tcp.Name != "bar" {
			t.Fatal("expected tcp check bar, got", tcp.Name)
		}
	} else {
		t.Fatal("expected checks[1] tcp")
	}
}

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
