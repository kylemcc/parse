package parse

import (
	"encoding/json"
	"net/url"
)

const HealthCheckEndPoint = "/health"

type healthCheckT struct {
}

func ServerHealthCheck() (map[string]string, error) {

	body, err := defaultClient.doRequest(&healthCheckT{})
	if err != nil {
		return map[string]string{"status": "fail"}, err
	}
	data := map[string]string{}
	if err := json.Unmarshal(body, &data); err != nil {
		return map[string]string{"status": "fail"}, err
	}
	return map[string]string{"status": "ok"}, nil
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
