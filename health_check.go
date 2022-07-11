package parse

import (
	"encoding/json"
	"net/url"
)

const HealthCheckEndPoint = "/health"

type healthCheckT struct {
}

// To check if the server is up and running.
func ServerHealthCheck() (map[string]interface{}, error) {

	body, err := defaultClient.doRequest(&healthCheckT{})
	if err != nil {
		return nil, err
	}
	resp := map[string]interface{}{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (h *healthCheckT) method() string {
	return "GET"
}

func (h *healthCheckT) endpoint() (string, error) {

	u := url.URL{}
	u.Scheme = ParseScheme
	u.Host = parseHost
	u.Path = ParsePath + HealthCheckEndPoint
	return u.String(), nil
}

func (h *healthCheckT) body() (string, error) {
	return "", nil
}

func (h *healthCheckT) useMasterKey() bool {
	return false
}

func (h *healthCheckT) session() *sessionT {
	return nil
}

func (h *healthCheckT) contentType() string {
	return "application/json"
}
