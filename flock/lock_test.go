package flock

import (
	"os"
	"testing"
)

func TestLockIsLocking(t *testing.T) {
	path := "test.lock"
	l1, err := Open(path)
	if err != nil {
		t.Fatal("l1 error", err)
	}
	defer l1.Close()

	// Test that the second lock cannot be acquired
	l2, err := Open(path)
	if l2 != nil {
		t.Fatal("l2 expected to be nil")
	}
	if err == nil || err.Error() != "resource temporarily unavailable" {
		t.Fatal("unexpected error", err)
	}

}

func TestLockIsNotLeavingFilesAround(t *testing.T) {
	path := "test.lock"
	l1, err := Open(path)
	if err != nil {
		t.Fatal("l1 error", err)
	}
	err = l1.Close()
	if err != nil {
		t.Fatal("l1 close error", err)
	}

	// The file shouldn't exist after that
	_, err = os.Stat(path)
	if err == nil || err.Error() != "stat test.lock: no such file or directory" {
		t.Fatal("unexpected error", err)
	}
}

func TestLockIsWorkingWithUnacquiredFile(t *testing.T) {
	path := "test.lock"
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("woot"))
	f.Close()

	l1, err := Open(path)
	if err != nil {
		t.Fatal("l1 error", err)
	}
	err = l1.Close()
	if err != nil {
		t.Fatal("l1 close error", err)
	}
}
