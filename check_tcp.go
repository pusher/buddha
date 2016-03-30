package buddha

import (
	"fmt"
	"net"
	"time"

	"github.com/pusher/buddha/log"
)

// establish tcp session to health check
type CheckTCP struct {
	// name of check in logs
	Name string `json:"name"`

	// host:port of tcp server under test
	Addr string `json:"addr"`
}

func (c CheckTCP) Validate() error {
	if len(c.Addr) == 0 {
		return fmt.Errorf("expected addr host:port for tcp check")
	}

	return nil
}

func (c CheckTCP) Execute(timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", c.Addr, timeout)
	if err != nil {
		log.Println(log.LevelInfo, "TCP connection failed: %s", err)
		return CheckFalse(fmt.Sprintf("TCP connection failed: %s", err))
	}
	defer conn.Close()

	return nil
}

func (c CheckTCP) String() string {
	return c.Name
}
