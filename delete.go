package parse

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"reflect"
)

// Delete the instance of the type represented by v from the Parse database. If
// useMasteKey=true, the Master Key will be used for the deletion request.
func Delete(v interface{}, useMasterKey bool) error {
	return _delete(v, useMasterKey, nil)
}

func _delete(v interface{}, useMasterKey bool, currentSession *sessionT) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a non-nil pointer")
	}

	_, err := defaultClient.doRequest(&deleteT{inst: v, shouldUseMasterKey: useMasterKey, currentSession: currentSession})
	return err
}

type deleteT struct {
	inst               interface{}
	shouldUseMasterKey bool
	currentSession     *sessionT
}

func (d *deleteT) method() string {
	return "DELETE"
}

func (d *deleteT) endpoint() (string, error) {
	var id string
	rv := reflect.ValueOf(d.inst)
	rvi := reflect.Indirect(rv)
	if f := rvi.FieldByName("Id"); f.IsValid() {
		if s, ok := f.Interface().(string); ok {
			id = s
		} else {
			return "", fmt.Errorf("Id field should be a string, received type %s", f.Type())
		}
	} else {
		return "", fmt.Errorf("can not delete value - type has no Id field")
	}

	p := getEndpointBase(d.inst)
	u := url.URL{}
	u.Scheme = "https"
	u.Host = parseHost
	u.Path = path.Join(p, id)

	return u.String(), nil
}

func (d *deleteT) body() (string, error) {
	return "", nil
}

func (d *deleteT) useMasterKey() bool {
	return d.shouldUseMasterKey
}

func (d *deleteT) session() *sessionT {
	return d.currentSession
}

func (d *deleteT) contentType() string {
	return "application/x-www-form-urlencoded"
}
