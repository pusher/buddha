package buddha

import (
	"os"
	"testing"
)

func TestJobsSelect(t *testing.T) {
	j := Jobs{&Job{Name: "foo"}, &Job{Name: "bar"}}

	j, missing := j.Select([]string{"bar"})

	if len(missing) > 0 {
		t.Fatal("expected no job missing", missing)
	} else if len(j) != 1 {
		t.Fatal("expected 1 job, got", len(j))
	} else if j[0].Name != "bar" {
		t.Fatal("expected job[0] bar, got", j[0].Name)
	}
}

func TestJobsSelectOrder(t *testing.T) {
	j := Jobs{&Job{Name: "foo"}, &Job{Name: "bar"}}

	j, missing := j.Select([]string{"foo", "bar"})

	if len(missing) > 0 {
		t.Fatal("expected no job missing", missing)
	} else if len(j) != 2 {
		t.Fatal("expected 2 job, got", len(j))
	} else if j[0].Name != "foo" {
		t.Fatal("expected job[0] foo, got", j[0].Name)
	} else if j[1].Name != "bar" {
		t.Fatal("expected job[1] bar, got", j[1].Name)
	}
}

func TestJobsSelectMissing(t *testing.T) {
	j := Jobs{&Job{Name: "foo"}, &Job{Name: "bar"}}

	j, missing := j.Select([]string{"foo", "something-else"})

	if len(missing) != 1 {
		t.Fatal("exected 1 missing job", j, missing)
	} else if missing[0] != "something-else" {
		t.Fatal("expected missing job to be 'something-else")
	} else if len(j) != 1 {
		t.Fatal("expected 1 job", j)
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
