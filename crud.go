package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"reflect"
	"strings"
)

type Update interface {
	Set(f string, v interface{}) Update
	Increment(f string, v interface{}) Update
	Delete(f string) Update
	UseMasterKey() Update
	Execute() error
}

type updateT struct {
	inst               interface{}
	values             map[string]interface{}
	shouldUseMasterKey bool
	currentSession     *sessionT
}

func NewUpdate(v interface{}) (Update, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, errors.New("v must be a non-nil pointer")
	}

	return &updateT{
		inst:   v,
		values: map[string]interface{}{},
	}, nil
}

func (u *updateT) Set(f string, v interface{}) Update {
	u.values[f] = v
	return u
}

func (u *updateT) Increment(f string, v interface{}) Update {
	u.values[f] = map[string]interface{}{
		"__op":   "Increment",
		"amount": v,
	}
	return u
}

func (u *updateT) Delete(f string) Update {
	u.values[f] = map[string]interface{}{
		"__op": "Delete",
	}
	return u
}

func (u *updateT) Execute() error {
	rv := reflect.ValueOf(u.inst)
	rvi := reflect.Indirect(rv)
	fieldMap := getFieldNameMap(rv)

	for k, v := range u.values {
		var fname string
		if fn, ok := fieldMap[k]; ok {
			fname = fn
		} else {
			fname = k
		}

		fname = strings.Title(fname)

		if fv := rvi.FieldByName(fname); fv.IsValid() {
			dv := reflect.ValueOf(v)
			dvi := reflect.Indirect(dv)
			fvi := reflect.Indirect(fv)
			fvi.Set(dvi)
		}
	}
	return defaultClient.doRequest(u, u.inst)
}

func (u *updateT) UseMasterKey() Update {
	u.shouldUseMasterKey = true
	return u
}

func (u *updateT) method() string {
	return "PUT"
}

func (u *updateT) endpoint() (string, error) {
	_url := url.URL{}
	p := getEndpointBase(u.inst)

	rv := reflect.ValueOf(u.inst)
	if f := rv.FieldByName("Id"); f.IsValid() {
		if s, ok := f.Interface().(string); ok {
			p = path.Join(p, s)
		} else {
			return "", fmt.Errorf("Id field should be a string, received type %s", f.Type())
		}
	} else {
		return "", fmt.Errorf("can not update value - type has no Id field")
	}

	_url.Scheme = "https"
	_url.Host = "api.parse.com"
	_url.Path = p

	return _url.String(), nil
}

func (u *updateT) body() (string, error) {
	b, err := json.Marshal(u.values)
	if err != nil {
		return "", err
	}

	return url.QueryEscape(string(b)), nil
}

func (u *updateT) useMasterKey() bool {
	return u.shouldUseMasterKey
}

func (u *updateT) session() *sessionT {
	return u.currentSession
}

func UpdateModel(v interface{}) error {
	return nil
}
