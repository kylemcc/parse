package parse

import (
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"time"
)

// An interface for custom Parse types. Contains a single method:
//
// ClassName() - returns a string containing the class name as it appears in your
// Parse database.
//
// Implement this interface if your class name does not match your struct
// name. If this class is not implemented, the name of the struct will
// be used when interacting with the Parse API
type iClassName interface {
	ClassName() string
}

// An interface for custom Parse types to override the endpoint used for querying.
//
// Contains a single method:
//
// Endpoint() - returns the endpoint to use when querying the Parse REST API.
//
// If this method is not implented, the endpoint is constructed as follows:
//
// /classes/{ClassName} - where {ClassName} is the name of the struct or the value returned by the ClassName
// method if implemented
type iParseEp interface {
	Endpoint() string
}

// A base type containing fields common to all Parse types
//
// Embed this struct in custom types to avoid having to declare
// these fields everywhere.
type Base struct {
	Id        string     `parse:"objectId"`
	CreatedAt *time.Time `parse:"-"`
	UpdatedAt *time.Time `parse:"-"`
	ACL       *ACL
	Extra     map[string]interface{} `parse:"-"`
}

// Represents the built-in Parse "User" class. Embed this type in a custom
// type containing any custom fields. When fetching user objects, any retrieved
// fields with no matching struct field will be stored in User.Extra (map[string]interface{})
type User struct {
	Base
}

func (u *User) ClassName() string {
	return "_User"
}

func (u *User) Endpoint() string {
	return "users"
}

// Represents the built-in Parse "Installation" class. Embed this type in a custom
// type containing any custom fields. When fetching user objects, any retrieved
// fields with no matching struct field will be stored in User.Extra (map[string]interface{})
type Installation struct {
	Base
}

func (i *Installation) ClassName() string {
	return "_Installation"
}

func (i *Installation) Endpoint() string {
	return "installations"
}

type ACL struct {
}

// Represents the Parse GeoPoint type
type GeoPoint struct {
	Latitude  float64
	Longitude float64
}

func (g GeoPoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Type      string  `json:"__type"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}{
		"GeoPoint",
		g.Latitude,
		g.Longitude,
	})
}

func (g *GeoPoint) UnmarshalJSON(b []byte) error {
	s := struct {
		Type      string  `json:"__type"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}{}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	if s.Type != "GeoPoint" {
		return fmt.Errorf("cannot unmarshal type %s to type GeoPoint", s.Type)
	}

	g.Latitude = s.Latitude
	g.Longitude = s.Longitude
	return nil
}

// Represents the Parse File type
type File struct {
}

// Represents a Parse Pointer type. When querying, creating, or updating
// objects, any struct types will be automatically converted to and from Pointer
// types as required. Direct use of this type should not be necessary
type Pointer struct {
	Id        string
	ClassName string
}

func (p Pointer) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Type      string `json:"__type"`
		ClassName string `json:"className"`
		Id        string `json:"objectId"`
	}{
		"Pointer",
		p.ClassName,
		p.Id,
	})
}

// Represents the Parse Date type. Values of type time.Time will
// automatically converted to a Date type when constructing queries
// or creating objects. The inverse is true for retrieving objects.
// Direct use of this type should not be necessary
type Date time.Time

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Type string `json:"__type"`
		Iso  string `json:"iso"`
	}{
		"Date",
		time.Time(d).In(time.UTC).Format("2006-01-02T15:04:05.000Z"),
	})
}

func (d *Date) UnmarshalJSON(b []byte) error {
	s := struct {
		Type string `json:"__type"`
		Iso  string `json:"iso"`
	}{}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	if s.Type != "Date" {
		return fmt.Errorf("cannot unmarshal type %s to type Date", s.Type)
	}

	t, err := time.Parse(s.Iso, "2006-01-02T15:04:05.000Z")
	if err != nil {
		return err
	}

	*d = Date(t)
	return nil
}

func getEndpointBase(v interface{}) string {
	var p string
	var inst interface{}

	rt := reflect.TypeOf(v)
	rt = rt.Elem()
	if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
		rte := rt.Elem()
		var rv reflect.Value
		if rte.Kind() == reflect.Ptr {
			rv = reflect.New(rte.Elem())
		} else {
			rv = reflect.New(rte)
		}
		inst = rv.Interface()
	} else {
		inst = v
	}

	if iv, ok := inst.(iParseEp); ok {
		p = iv.Endpoint()
	} else {
		var cname string
		if v, ok := inst.(iClassName); ok {
			cname = v.ClassName()
		} else {
			t := reflect.TypeOf(inst)
			cname = t.Elem().Name()
		}
		p = path.Join("classes", cname)
	}

	p = path.Join(ParseVersion, p)
	return p
}
