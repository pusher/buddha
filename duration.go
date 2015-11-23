package buddha

import (
	"fmt"
	"time"
)

// replacement for time.Duration to satisfy encoding/json Unmarshaler interface
type Duration time.Duration

func (d Duration) String() string {
	return d.Duration().String()
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d *Duration) UnmarshalJSON(p []byte) error {
	if len(p) < 2 || p[0] != '"' || p[len(p)-1] != '"' {
		return fmt.Errorf("invalid duration string: %s", string(p))
	}

	dur, err := time.ParseDuration(string(p[1 : len(p)-1]))
	if err != nil {
		return err
	}

	*d = Duration(dur)

	return nil
}
