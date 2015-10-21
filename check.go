package buddha

import (
	"time"
)

type Checks []Check

type Check interface {
	// identifier string for check
	String() string

	// validate check options for correctness
	Validate() error

	// execute health check with timeout
	Execute(time.Duration) error
}
