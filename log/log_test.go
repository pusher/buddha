package log

import (
	"bytes"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	l := New(nil)
	if l.out != os.Stdout {
		t.Fatal("expected out os.Stdout, got", l.out)
	}
}

func TestLoggerSetOutput(t *testing.T) {
	l := New(nil)
	l.SetOutput(nil)
	if l.out != nil {
		t.Fatal("expected out nil, got", l.out)
	}
}

func TestLoggerPrint(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	l := New(buf)

	n, err := l.Print(2, "hello %s", "world")
	if err != nil {
		t.Fatal("unexpected error:", err)
	} else if n != 15 {
		t.Fatal("expected 15 written, got", n)
	}

	if s := buf.String(); s != "--> hello world" {
		t.Fatalf("expected '--> hello world', got '%s'", s)
	}
}

func TestLoggerPrintln(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	l := New(buf)

	n, err := l.Println(2, "hello %s", "world")
	if err != nil {
		t.Fatal("unexpected error:", err)
	} else if n != 17 {
		t.Fatal("expected 17 written, got", n)
	}

	if s := buf.String(); s != "--> hello world\r\n" {
		t.Fatalf("expected '--> hello world\r\n', got '%s'", s)
	}
}

func TestLoggerFormat(t *testing.T) {
	l := New(nil)

	if s := l.format(0, "foo"); s != "foo" {
		t.Fatalf("expected 'foo', got '%s'", s)
	} else if s := l.format(1, "foo"); s != "    foo" {
		t.Fatalf("expected '    foo', got '%s'", s)
	} else if s := l.format(2, "foo"); s != "--> foo" {
		t.Fatalf("expected '--> foo', got '%s'", s)
	} else if s := l.format(3, "foo"); s != "==> foo" {
		t.Fatalf("expected '==> foo', got '%s'", s)
	} else if s := l.format(4, "foo"); s != "!!! foo" {
		t.Fatalf("expected '!!! foo', got '%s'", s)
	}
}
