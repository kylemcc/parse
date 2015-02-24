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

	//Set the field specified by f to the value of v
	Set(f string, v interface{}) Update

	// Increment the field specified by f by the amount specified by v.
	// v should be a numeric type
	Increment(f string, v interface{}) Update

	// Delete the field specified by f from the instance being updated
	Delete(f string) Update

	// Append the values provided to the Array field specified by f. This operation
	// is atomic
	Add(f string, vs ...interface{}) Update

	// Add any values provided that were not alread present to the Array field
	// specified by f. This operation is atomic
	AddUnique(f string, vs ...interface{}) Update

	// Remove the provided values from the array field specified by f
	Remove(f string, vs ...interface{}) Update

	// Update the ACL on the given object
	SetACL(a ACL) Update

	// Use the Master Key for this update request
	UseMasterKey() Update

	// Execute this update. This method also updates the proper fields
	// on the provided value with their repective new values
	Execute() error

	requestT
}

type updateT struct {
	inst               interface{}
	values             map[string]updateOpT
	shouldUseMasterKey bool
	currentSession     *sessionT
}

// Create a new update request for the Parse object represented by v.
//
// Note: v should be a pointer to a struct whose name represents a Parse class,
// or that implements the ClassName method
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
	u.values[f] = updateOpT{UpdateType: opSet, Value: encodeForRequest(v)}
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

func (u *updateT) SetACL(a ACL) Update {
	u.values["ACL"] = updateOpT{UpdateType: opSet, Value: a}
	return u
}

func (u *updateT) Execute() (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("error executing update: %v", r)
			}
		}
	}()

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
				if fv.Kind() == reflect.Ptr && fv.IsNil() && v.Value != nil {
					fv.Set(reflect.New(fv.Type().Elem()))
				}

				var tmp reflect.Value
				if fv.Kind() == reflect.Ptr {
					if v.Value == nil {
						tmp = fv.Addr()
					} else {
						tmp = fv
					}
				} else {
					tmp = fv.Addr()
				}
				if err := populateValue(tmp.Interface(), v.Value); err != nil {
					return err
				}
			case opIncr:
				switch fvi.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if dvi.Type().ConvertibleTo(fvi.Type()) {
						current := fvi.Int()
						amount := dvi.Convert(fvi.Type()).Int()
						current += amount
						fvi.Set(reflect.ValueOf(current).Convert(fvi.Type()))
					}
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if dvi.Type().ConvertibleTo(fvi.Type()) {
						current := fvi.Uint()
						amount := dvi.Convert(fvi.Type()).Uint()
						current += amount
						fvi.Set(reflect.ValueOf(current).Convert(fvi.Type()))
					}
				case reflect.Float32, reflect.Float64:
					if dvi.Type().ConvertibleTo(fvi.Type()) {
						current := fvi.Float()
						amount := dvi.Convert(fvi.Type()).Float()
						current += amount
						fvi.Set(reflect.ValueOf(current).Convert(fvi.Type()))
					}
				}
			case opDelete:
				fv.Set(reflect.Zero(fv.Type()))
			}
		}
	}
	if b, err := defaultClient.doRequest(u); err != nil {
		return err
	} else {
		return handleResponse(b, u.inst)
	}
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
	_url.Host = parseHost
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

func LinkFacebookAccount(u *User, a *FacebookAuthData) error {
	if u.Id == "" {
		return errors.New("user Id field must not be empty")
	}

	up, _ := NewUpdate(u)
	up.Set("authData", AuthData{Facebook: a})
	up.UseMasterKey()
	return up.Execute()
}
