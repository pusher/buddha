package buddha

import (
	"os"
	"testing"
)

func TestOpen(t *testing.T) {
	file, err := os.OpenFile("example/reload_app_servers.json", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer file.Close()

	jobs, err := Open(file)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(*jobs) != 1 {
		t.Fatal("expected 1 job, got", len(*jobs))
	}
}

func TestOpenFile(t *testing.T) {
	jobs, err := OpenFile("example/reload_app_servers.json")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(*jobs) != 1 {
		t.Fatal("expected 1 job, got", len(*jobs))
	}
}
