package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"reflect"
)

type updateTypeT int

const (
	opSet updateTypeT = iota
	opIncr
	opDelete
	opAdd
	opAddUnique
	opRemove
	opAddRelation
	opRemoveRelation
)

func (u updateTypeT) String() string {
	switch u {
	case opSet:
		return "Set"
	case opIncr:
		return "Increment"
	case opDelete:
		return "Delete"
	case opAdd:
		return "Add"
	case opAddUnique:
		return "AddUnique"
	case opRemove:
		return "Remove"
	case opAddRelation:
		return "AddRelation"
	case opRemoveRelation:
		return "RemoveRelation"
	}

	return "Unknown"
}

func (u updateTypeT) argKey() string {
	switch u {
	case opIncr:
		return "amount"
	case opAdd, opAddUnique, opRemove, opAddRelation, opRemoveRelation:
		return "objects"
	}

	return "unknown"
}

type updateOpT struct {
	UpdateType updateTypeT
	Value      interface{}
}

func (u updateOpT) MarshalJSON() ([]byte, error) {
	switch u.UpdateType {
	case opSet:
		return json.Marshal(u.Value)
	case opDelete:
		return json.Marshal(map[string]interface{}{
			"__op": u.UpdateType.String(),
		})
	default:
		return json.Marshal(map[string]interface{}{
			"__op":                u.UpdateType.String(),
			u.UpdateType.argKey(): u.Value,
		})
	}
}

type Update interface {
	Set(f string, v interface{}) Update
	Increment(f string, v interface{}) Update
	Delete(f string) Update
	Add(f string, vs ...interface{}) Update
	AddUnique(f string, vs ...interface{}) Update
	Remove(f string, vs ...interface{}) Update
	UseMasterKey() Update
	Execute() error
}

type updateT struct {
	inst               interface{}
	values             map[string]updateOpT
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
		values: map[string]updateOpT{},
	}, nil
}

func (u *updateT) Set(f string, v interface{}) Update {
	u.values[f] = updateOpT{UpdateType: opSet, Value: v}
	return u
}

func (u *updateT) Increment(f string, v interface{}) Update {
	u.values[f] = updateOpT{UpdateType: opIncr, Value: v}
	return u
}

func (u *updateT) Delete(f string) Update {
	u.values[f] = updateOpT{UpdateType: opDelete}
	return u
}

func (u *updateT) Add(f string, vs ...interface{}) Update {
	u.values[f] = updateOpT{UpdateType: opAdd, Value: vs}
	return u
}

func (u *updateT) AddUnique(f string, vs ...interface{}) Update {
	u.values[f] = updateOpT{UpdateType: opAddUnique, Value: vs}
	return u
}

func (u *updateT) Remove(f string, vs ...interface{}) Update {
	u.values[f] = updateOpT{UpdateType: opRemove, Value: vs}
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

		fname = firstToUpper(fname)

		dv := reflect.ValueOf(v.Value)
		dvi := reflect.Indirect(dv)

		if fv := rvi.FieldByName(fname); fv.IsValid() {
			fvi := reflect.Indirect(fv)

			switch v.UpdateType {
			case opSet:
				fvi.Set(dvi)
			case opIncr:
				switch fvi.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if dvi.Type().AssignableTo(fvi.Type()) {
						current := fvi.Int()
						amount := dvi.Int()
						current += amount
						fvi.Set(reflect.ValueOf(current).Convert(fvi.Type()))
					}
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if dvi.Type().AssignableTo(fvi.Type()) {
						current := fvi.Uint()
						amount := dvi.Uint()
						current += amount
						fvi.Set(reflect.ValueOf(current))
					}
				case reflect.Float32, reflect.Float64:
					if dvi.Type().AssignableTo(fvi.Type()) {
						current := fvi.Float()
						amount := dvi.Float()
						current += amount
						fvi.Set(reflect.ValueOf(current))
					}
				}
			case opDelete:
				fv.Set(reflect.Zero(fv.Type()))
			}
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
	rvi := reflect.Indirect(rv)
	if f := rvi.FieldByName("Id"); f.IsValid() {
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

	return string(b), nil
}

func (u *updateT) useMasterKey() bool {
	return u.shouldUseMasterKey
}

func (u *updateT) session() *sessionT {
	return u.currentSession
}

func (u *updateT) contentType() string {
	return "application/json"
}
