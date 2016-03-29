package buddha

import (
	"net"
	"testing"
	"time"

	"github.com/pusher/buddha/tcptest"
)

func TestCheckTCPValidate(t *testing.T) {
	c1 := CheckTCP{}
	if err := c1.Validate(); err == nil {
		t.Fatal("expected error, got nil")
	}

	c2 := CheckTCP{Addr: "127.0.0.1:8080"}
	if err := c2.Validate(); err != nil {
		t.Fatal("expected nil, got", err)
	}
}

func TestCheckTCPExecute(t *testing.T) {
	ts := tcptest.NewServer(func(conn net.Conn) {
		defer conn.Close()
		// do nothing
	})
	defer ts.Close()

	c := CheckTCP{Addr: ts.Addr.String()}
	_, err := c.Execute(1 * time.Second)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func TestCheckTCPString(t *testing.T) {
	c := CheckTCP{Name: "foo"}

	if s := c.String(); s != "foo" {
		t.Fatal("expected string foo, got", s)
	}
}
