package parse

import (
	"encoding/json"
	"errors"
	"net/url"
	"reflect"
)

type createT struct {
	v                  interface{}
	shouldUseMasterKey bool
	currentSession     *sessionT
}

func (c *createT) method() string {
	return "POST"
}

func (c *createT) endpoint() (string, error) {
	p := getEndpointBase(c.v)
	u := url.URL{}
	u.Scheme = "https"
	u.Host = parseHost
	u.Path = p

	return u.String(), nil
}

func (c *createT) body() (string, error) {
	payload := map[string]interface{}{}

	rv := reflect.ValueOf(c.v)
	rvi := reflect.Indirect(rv)
	rt := rvi.Type()
	fields := getFields(rt)

	for _, f := range fields {
		var t string
		if t = f.Tag.Get("parse"); t == "-" || t == "objectId" || f.Name == "Id" {
			continue
		}
		var fname string
		if t != "" {
			fname = t
		} else {
			fname = firstToLower(f.Name)
		}

		if fv := rvi.FieldByName(f.Name); fv.IsValid() {
			if fname == "ACL" && fv.IsNil() {
				continue
			}
			payload[fname] = fv.Interface()
		}
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (c *createT) useMasterKey() bool {
	return c.shouldUseMasterKey
}

func (c *createT) session() *sessionT {
	return c.currentSession
}

func (c *createT) contentType() string {
	return "application/json"
}

// Save a new instance of the type pointed to by v to the Parse database. If
// useMasteKey=true, the Master Key will be used for the creation request. On a
// successful request, the CreatedAt field will be set on v.
//
// Note: v should be a pointer to a struct whose name represents a Parse class,
// or that implements the ClassName method
func Create(v interface{}, useMasterKey bool) error {
	return create(v, useMasterKey, nil)
}

func create(v interface{}, useMasterKey bool, currentSession *sessionT) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a non-nil pointer")
	}

	cr := &createT{
		v:                  v,
		shouldUseMasterKey: useMasterKey,
		currentSession:     currentSession,
	}
	return defaultClient.doRequest(cr, v)
}
