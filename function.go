package parse

import (
	"encoding/json"
	"errors"
	"net/url"
	"path"
	"reflect"
)

type Params map[string]interface{}

func CallFunction(name string, params Params, resp interface{}) error {
	return callFn(name, params, resp, nil)
}

type callFnT struct {
	name           string
	params         Params
	currentSession *sessionT
}

func (c *callFnT) method() string {
	return "POST"
}

func (c *callFnT) endpoint() (string, error) {
	p := path.Join(ParseVersion, "functions", c.name)
	u := url.URL{}
	u.Scheme = "https"
	u.Host = parseHost
	u.Path = p

	return u.String(), nil
}

func (c *callFnT) body() (string, error) {
	b, err := json.Marshal(c.params)
	return string(b), err
}

func (c *callFnT) useMasterKey() bool {
	return false
}

func (c *callFnT) session() *sessionT {
	return c.currentSession
}

func (c *callFnT) contentType() string {
	return "application/json"
}

type fnRespT struct {
	Result interface{} `parse:"result"`
}

func callFn(name string, params Params, resp interface{}, currentSession *sessionT) error {
	rv := reflect.ValueOf(resp)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("resp must be a non-nil pointer")
	}

	if params == nil {
		params = Params{}
	}

	cr := &callFnT{
		name:           name,
		params:         params,
		currentSession: currentSession,
	}
	if b, err := defaultClient.doRequest(cr); err != nil {
		return err
	} else {
		r := fnRespT{}
		if err := json.Unmarshal(b, &r); err != nil {
			return err
		}
		return populateValue(resp, r.Result)
	}
}
