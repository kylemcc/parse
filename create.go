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

	isUser   bool
	username string
	password string
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

	if c.isUser {
		payload["username"] = c.username
		payload["password"] = c.password
	}

	rv := reflect.ValueOf(c.v)
	rvi := reflect.Indirect(rv)
	rt := rvi.Type()
	fields := getFields(rt)

	for _, f := range fields {
		var name string
		var fv reflect.Value

		if n, o := parseTag(f.Tag.Get("parse")); n == "-" || n == "objectId" || f.Name == "Id" || f.Type == reflect.TypeOf(Base{}) {
			continue
		} else if fv = rvi.FieldByName(f.Name); !fv.IsValid() || o == "omitempty" && isEmptyValue(fv) {
			continue
		} else {
			name = n
		}

		var fname string
		if name != "" {
			fname = name
		} else {
			fname = firstToLower(f.Name)
		}

		if canBeNil(fv) && fv.IsNil() {
			payload[fname] = nil
		} else {
			payload[fname] = encodeForRequest(fv.Interface())
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

func Signup(username string, password string, user interface{}) error {
	cr := &createT{
		v:                  user,
		shouldUseMasterKey: false,
		currentSession:     nil,
		isUser:             true,
		username:           username,
		password:           password,
	}
	if b, err := defaultClient.doRequest(cr); err != nil {
		return err
	} else {
		return handleResponse(b, user)
	}
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
	if b, err := defaultClient.doRequest(cr); err != nil {
		return err
	} else {
		return handleResponse(b, v)
	}
}
