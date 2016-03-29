package buddha

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckHTTPValidate(t *testing.T) {
	c1 := CheckHTTP{}
	if err := c1.Validate(); err == nil {
		t.Fatal("expected error, got nil")
	}

	c2 := CheckHTTP{Path: "http://127.0.0.1:8080/health_check"}
	if err := c2.Validate(); err != nil {
		t.Fatal("expected nil, got", err)
	}
}

func TestCheckHTTPExecute(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer ts.Close()

	c := CheckHTTP{Path: ts.URL, Expect: []int{204}}
	_, err := c.Execute(1 * time.Second)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func TestCheckHTTPCheckStatusCode(t *testing.T) {
	c := CheckHTTP{Expect: []int{200}}

	if !c.checkStatusCode(200) {
		t.Fatal("unexpected 200 status code failure")
	} else if c.checkStatusCode(500) {
		t.Fatal("unexpected 500 status code pass")
	}
}

func TestCheckHTTPString(t *testing.T) {
	c := CheckHTTP{Name: "foo"}

	if s := c.String(); s != "foo" {
		t.Fatal("expected string foo, got", s)
	}
}
