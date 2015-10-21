package buddha

import (
	"fmt"
	"net"
	"time"
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
		return err
	}
	defer conn.Close()

	return nil
}

func (c CheckTCP) String() string {
	return c.Name
}
