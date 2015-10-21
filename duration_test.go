package buddha

import (
	"testing"
	"time"
)

func TestDurationString(t *testing.T) {
	d1 := Duration(1 * time.Second)
	if s := d1.String(); s != "1s" {
		t.Fatal("expected 1s, got", s)
	}

	d2 := Duration((1 * time.Hour) + (30 * time.Minute))
	if s := d2.String(); s != "1h30m0s" {
		t.Fatal("expected 1h30m0s, got", s)
	}
}

func TestDurationDuration(t *testing.T) {
	d := Duration(1 * time.Second)
	td := d.Duration()

	if s := td.String(); s != "1s" {
		t.Fatal("expected 1s, got", s)
	}
}

func TestDurationUnmarshalJSON(t *testing.T) {
	var d Duration
	err := d.UnmarshalJSON([]byte(`"1h30m0s"`))
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if s := d.String(); s != "1h30m0s" {
		t.Fatal("expected 1h30m0s, got", s)
	}
}
