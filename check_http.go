package buddha

import (
	"fmt"
	"net/http"
	"time"
)

// execute OPTIONS http request to health check
type CheckHTTP struct {
	// name of check in logs
	Name string `json:"name"`

	// http method to use
	// OPTIONS is the recommended
	Method string `json:"method"`

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
	if c.Method == "" {
		c.Method = "OPTIONS"
	}

	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest(c.Method, c.Path, nil)
	if err != nil {
		return fmt.Errorf("building http request failed %s", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return CheckFalse(fmt.Sprintf("HTTP request failed: %s", err))
	}
	defer res.Body.Close()

	if !c.checkStatusCode(res.StatusCode) {
		return CheckFalse(fmt.Sprintf("Unacceptable status code %d", res.StatusCode))
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
