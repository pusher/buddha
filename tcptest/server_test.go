package tcptest

import (
	"io/ioutil"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	ts := NewServer(func(conn net.Conn) {
		defer conn.Close()

		conn.Write([]byte("hello world\r\n"))
	})
	defer ts.Close()

	conn, err := net.Dial("tcp", ts.Addr.String())
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer conn.Close()

	data, err := ioutil.ReadAll(conn)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if string(data) != "hello world\r\n" {
		t.Fatalf("data mismatch: expected 'hello world', got '%s'", data)
	}
}
