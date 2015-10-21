tcptest
=======

tcptest implements a tcp server, similar in use to stdlib net/http/httptest, for the testing of client tcp connections.

[GoDoc](https://godoc.org/github.com/pusher/buddha/tcptest)

Example
-------

	package tcptest

	import (
		"io/ioutil"
		"net"
		"testing"

		"github.com/pusher/buddha/tcptest"
	)

	func TestServer(t *testing.T) {
		ts := tcptest.NewServer(func(conn net.Conn) {
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
