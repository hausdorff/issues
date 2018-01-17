package server

import (
	"net/http"
	"testing"
)

func TestParse(t *testing.T) {
	go Serve("8080")

	resp, err := http.Get("http://127.0.0.1:8080/api/v0.1/issues")
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Status code: %d", resp.StatusCode)
	}
}
