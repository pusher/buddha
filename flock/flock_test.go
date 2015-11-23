package flock

import (
	"os"
	"testing"
)

func TestLock(t *testing.T) {
	f, err := Lock("test/lock.pid")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer f.Close()

	_, err = os.Lstat("test/lock.pid")
	if os.IsNotExist(err) {
		t.Fatal("expected test/lock.pid to exist")
	}
}

func TestLockLocked(t *testing.T) {
	n, err := os.OpenFile("test/lock.pid", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	n.Close()
	defer os.Remove("test/lock.pid")

	_, err = Lock("test/lock.pid")
	if err != ErrLocked {
		t.Fatal("expected ErrLocked, got", err)
	}
}

func TestLockClose(t *testing.T) {
	f, err := Lock("test/lock.pid")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = f.Close()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	_, err = os.Lstat("test/lock.pid")
	if !os.IsNotExist(err) {
		t.Fatal("unexpected error:", err)
	}
}
