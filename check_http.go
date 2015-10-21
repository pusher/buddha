package buddha

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

// execute OPTIONS http request to health check
type CheckHTTP struct {
	// name of check in logs
	Name string `json:"name"`

	// url to issue health check
	Path string `json:"path"`

	// expected HTTP status codes
	Expect []int `json:"expect,omitempty"`
}

func (c CheckHTTP) Validate() error {
	if len(c.Path) == 0 {
		return fmt.Errorf("expected path URL for http check")
	}

	return nil
}

func (c CheckHTTP) Execute(timeout time.Duration) error {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, address string) (net.Conn, error) {
				return net.DialTimeout(network, address, timeout)
			},
		},
	}

	req, err := http.NewRequest("OPTIONS", c.Path, nil)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if !c.checkStatusCode(res.StatusCode) {
		return fmt.Errorf("unacceptable status code %d", res.StatusCode)
	}

	return nil
}

func (c CheckHTTP) String() string {
	return c.Name
}

func (c CheckHTTP) checkStatusCode(code int) bool {
	for _, i := range c.Expect {
		if i == code {
			return true
		}
	}
	return false
}
