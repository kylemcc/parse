package parse

import (
	"encoding/json"
	"fmt"
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
// Embed this struct in custom types
type Base struct {
	Id        string `parse:"objectId"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
	Acl       *ACL
	Extra     map[string]interface{}
}

// Represents the built-in Parse "User" class
type User struct {
	Base
}

func (u *User) ClassName() string {
	return "_User"
}

func (u *User) Endpoint() string {
	return "users"
}

// Represents the built-in Parse "Installation" class
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

type GeoPoint struct {
}

type File struct {
}

type Pointer struct {
	Id string
	ClassName string
}

func (p Pointer) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Type string `json:"__type"`
		ClassName string `json:"className"`
		Id string `json:"objectId"`
	}{
		"Pointer",
		p.ClassName,
		p.Id,
	})
}

type Date time.Time

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Type string `json:"__type"`
		Iso  string `json:"iso"`
	}{
		"Date",
		time.Time(d).Format(time.RFC3339),
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

	t, err := time.Parse(s.Iso, time.RFC3339Nano)
	if err != nil {
		return err
	}

	*d = Date(t)
	return nil
}
