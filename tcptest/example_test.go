package tcptest_test

import (
	"io/ioutil"
	"log"
	"net"

	"github.com/pusher/buddha/tcptest"
)

func ExampleServer() {
	ts := tcptest.NewServer(func(conn net.Conn) {
		defer conn.Close()

		conn.Write([]byte("hello world\r\n"))
	})
	defer ts.Close()

	conn, err := net.Dial("tcp", ts.Addr.String())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("data:", data)
}
