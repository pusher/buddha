// tcptest implements a tcp server, similar in use to stdlib net/http/httptest,
// for the testing of client tcp connections
package tcptest

import (
	"net"
)

// A Handler defines the interface the TCP test handlers must implement to
// accept connections.
type Handler func(net.Conn)

// A Server is a TCP server listening on a system-chosen port on the local
// loopback interface, for use in end-to-end TCP tests.
type Server struct {
	Addr net.Addr

	fn Handler
	ln *net.TCPListener
}

// NewServer starts and returns a new Server. The caller should call Close
// when finished, to shut it down
func NewServer(fn Handler) *Server {
	// listen on random system port >1024
	laddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:0")

	server, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		panic(err)
	}

	s := &Server{Addr: server.Addr(), fn: fn, ln: server}

	// launch background handler
	go s.serve()

	return s
}

// Close shuts down the server and blocks until all outstanding requests on this
// server have completed.
func (s Server) Close() error {
	return s.ln.Close()
}

// Serve accepts TCP connections and executes the Server handler in goroutine
func (s Server) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if operr, ok := err.(*net.OpError); ok { // >= go 1.5
				// gracefully handle socket closure (not error)
				if operr.Err.Error() == "use of closed network connection" {
					return
				}
			} else { // <= go 1.4
				if err.Error() == "use of closed network connection" {
					return
				}
			}

			panic(err)
		}

		go s.fn(conn)
	}
}
