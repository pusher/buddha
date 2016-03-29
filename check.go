package buddha

import (
	"encoding/json"
	"fmt"
	"time"
)

type Check interface {
	// identifier string for check
	String() string

	// validate check options for correctness
	Validate() error

	// execute health check with timeout
	// the bool result is whether the check passed or not
	// error will be set if there are any unexpected errors
	Execute(time.Duration) (bool, error)
}

type Checks []Check

// contextual unmarshaler into check types for interface
func (c *Checks) UnmarshalJSON(p []byte) error {
	var raw []json.RawMessage
	err := json.Unmarshal(p, &raw)
	if err != nil {
		return err
	}

	for _, r := range raw {
		var generic check
		err = json.Unmarshal(r, &generic)
		if err != nil {
			return err
		}

		switch generic.Type {
		case "http":
			var http CheckHTTP
			err := json.Unmarshal(r, &http)
			if err != nil {
				return err
			}

			*c = append(*c, http)

		case "tcp":
			var tcp CheckTCP
			err := json.Unmarshal(r, &tcp)
			if err != nil {
				return err
			}

			*c = append(*c, tcp)

		case "exec":
			var exec CheckExec
			err := json.Unmarshal(r, &exec)
			if err != nil {
				return err
			}

			*c = append(*c, exec)
		default:
			return fmt.Errorf("Unknown check type %s", generic.Type)
		}
	}

	return nil
}

type check struct {
	Type string `json:"type"`
}
