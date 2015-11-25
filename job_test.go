package buddha

import (
	"os"
	"testing"
)

func TestJobsFilter(t *testing.T) {
	j := Jobs{&Job{Name: "foo"}, &Job{Name: "bar"}}

	j = j.Filter([]string{"bar"})

	if len(j) != 1 {
		t.Fatal("expected 1 job, got", len(j))
	} else if j[0].Name != "bar" {
		t.Fatal("expected job[0] bar, got", j[0].Name)
	}
}

func TestJobsFilterOrder(t *testing.T) {
	j := Jobs{&Job{Name: "foo"}, &Job{Name: "bar"}}

	j = j.Filter([]string{"foo", "bar"})

	if len(j) != 2 {
		t.Fatal("expected 2 job, got", len(j))
	} else if j[0].Name != "foo" {
		t.Fatal("expected job[0] foo, got", j[0].Name)
	} else if j[1].Name != "bar" {
		t.Fatal("expected job[1] bar, got", j[1].Name)
	}
}

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

	if len(jobs) != 1 {
		t.Fatal("expected 1 job, got", len(jobs))
	}
}

func TestOpenFile(t *testing.T) {
	jobs, err := OpenFile("example/reload_app_servers.json")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if l := len(jobs); l != 1 {
		t.Fatal("expected 1 job, got", l)
	}
}

func TestOpenDir(t *testing.T) {
	jobs, err := OpenDir("example/")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if l := len(jobs); l != 2 {
		t.Fatal("expected 2 jobs, got", l)
	}
}
