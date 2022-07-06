package parse

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestServerHealthCheckStatusIsOk(t *testing.T) {

	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"status":"ok"}`)
	})
	defer teardownTestServer()

	expected := map[string]string{"status": "ok"}
	result, err := ServerHealthCheck()
	if err != nil {
		t.Error("Error must be nil while server status is ok!")
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ServerHealthCheck result is not formatted!")
	}
}

func TestServerHealthCheckStatusIsNotOk(t *testing.T) {

	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
		fmt.Fprintf(w, "")
	})
	defer teardownTestServer()

	expected := map[string]string{"status": "fail"}
	result, err := ServerHealthCheck()
	if err == nil {
		t.Error("Error must be available while server status is not ok!")
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ServerHealthCheck result is not formatted!")
	}
}
