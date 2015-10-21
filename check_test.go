package buddha

import (
	"encoding/json"
	"testing"
)

var testChecks = []byte(`[
  {"type": "http", "name": "http_8081", "path": "http://127.0.0.1:8081/health_check", "expect": [200]},
  {"type": "tcp", "name": "ws_8082", "addr": "127.0.0.1:8082"}
]`)

func TestChecksUnmarshalJSON(t *testing.T) {
	var checks Checks
	err := json.Unmarshal(testChecks, &checks)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if l := len(checks); l != 2 {
		t.Fatal("expected 2 checks, got", l)
	} else if s := checks[0].String(); s != "http_8081" {
		t.Fatal("expected checks[0] http_8081, got", s)
	} else if s := checks[1].String(); s != "ws_8082" {
		t.Fatal("expected checks[1] ws_8082", s)
	}
}
