package parse

import (
	"encoding/json"
	"fmt"
	"net/url"
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
	Id        string                 `parse:"objectId"`
	CreatedAt time.Time              `parse:"-"`
	UpdatedAt time.Time              `parse:"-"`
	ACL       ACL                    `parse:"ACL,omitempty"`
	Extra     map[string]interface{} `parse:"-"`
}

// Represents the built-in Parse "User" class. Embed this type in a custom
// type containing any custom fields. When fetching user objects, any retrieved
// fields with no matching struct field will be stored in User.Extra (map[string]interface{})
type User struct {
	Base
	Username      string
	Email         string
	EmailVerified bool
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

type ACL interface {
	// Returns whether public read access is enabled on this ACL
	PublicReadAccess() bool

	// Returns whether public write access is enabled on this ACL
	PublicWriteAccess() bool

	// Returns whether read access is enabled on this ACL for the
	// given role
	RoleReadAccess(role string) bool

	// Returns whether write access is enabled on this ACL for the
	// given role
	RoleWriteAccess(role string) bool

	// Returns whether read access is enabled on this ACL for the
	// given user
	ReadAccess(userId string) bool

	// Returns whether write access is enabled on this ACL for the
	// given user
	WriteAccess(userId string) bool

	// Allow the object to which this ACL is attached be read
	// by anyone
	SetPublicReadAccess(allowed bool) ACL

	// Allow the object to which this ACL is attached to be
	// updated by anyone
	SetPublicWriteAccess(allowed bool) ACL

	// Allow the object to which this ACL is attached to be
	// read by the provided role
	SetRoleReadAccess(role string, allowed bool) ACL

	// Allow the object to which this ACL is attached to be
	// updated by the provided role
	SetRoleWriteAccess(role string, allowed bool) ACL

	// Allow the object to which this ACL is attached to be
	// read by the provided user
	SetReadAccess(userId string, allowed bool) ACL

	// Allow the object to which this ACL is attached to be
	// updated by the provided user
	SetWriteAccess(userId string, allowed bool) ACL
}

type aclT struct {
	publicReadAccess  bool
	publicWriteAccess bool

	write map[string]bool
	read  map[string]bool
}

func NewACL() ACL {
	return &aclT{
		write: map[string]bool{},
		read:  map[string]bool{},
	}
}

func (a *aclT) PublicReadAccess() bool {
	return a.publicReadAccess
}

func (a *aclT) PublicWriteAccess() bool {
	return a.publicWriteAccess
}

func (a *aclT) RoleReadAccess(role string) bool {
	if tmp, ok := a.read["role:"+role]; ok {
		return tmp
	}
	return false
}

func (a *aclT) RoleWriteAccess(role string) bool {
	if tmp, ok := a.write["role:"+role]; ok {
		return tmp
	}
	return false
}

func (a *aclT) ReadAccess(userId string) bool {
	if tmp, ok := a.read[userId]; ok {
		return tmp
	}
	return false
}

func (a *aclT) WriteAccess(userId string) bool {
	if tmp, ok := a.write[userId]; ok {
		return tmp
	}
	return false
}

func (a *aclT) SetPublicReadAccess(allowed bool) ACL {
	a.publicReadAccess = allowed
	return a
}

func (a *aclT) SetPublicWriteAccess(allowed bool) ACL {
	a.publicWriteAccess = allowed
	return a
}

func (a *aclT) SetReadAccess(userId string, allowed bool) ACL {
	a.read[userId] = allowed
	return a
}

func (a *aclT) SetWriteAccess(userId string, allowed bool) ACL {
	a.write[userId] = allowed
	return a
}

func (a *aclT) SetRoleReadAccess(role string, allowed bool) ACL {
	a.read["role:"+role] = allowed
	return a
}

func (a *aclT) SetRoleWriteAccess(role string, allowed bool) ACL {
	a.write["role:"+role] = allowed
	return a
}

func (a *aclT) MarshalJSON() ([]byte, error) {
	m := map[string]map[string]bool{}

	for k, v := range a.read {
		if v {
			m[k] = map[string]bool{
				"read": v,
			}
		}
	}

	for k, v := range a.write {
		if v {
			if p, ok := m[k]; ok {
				p["write"] = v
			} else {
				m[k] = map[string]bool{
					"write": v,
				}
			}
		}
	}

	if a.publicReadAccess {
		m["*"] = map[string]bool{
			"read": true,
		}
	}

	if a.publicWriteAccess {
		if p, ok := m["*"]; !ok {
			m["*"] = map[string]bool{
				"write": true,
			}
		} else {
			p["write"] = true
		}
	}

	return json.Marshal(m)
}

func (a *aclT) UnmarshalJSON(b []byte) error {
	m := map[string]map[string]bool{}

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	if a.read == nil {
		a.read = map[string]bool{}
	}

	if a.write == nil {
		a.write = map[string]bool{}
	}

	for k, v := range m {
		if k == "*" {
			if w, ok := v["write"]; w && ok {
				a.publicWriteAccess = true
			}
			if r, ok := v["read"]; r && ok {
				a.publicReadAccess = true
			}
		} else {
			if w, ok := v["write"]; w && ok {
				a.write[k] = true
			}
			if r, ok := v["read"]; r && ok {
				a.read[k] = true
			}
		}
	}
	return nil
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

func getClassName(v interface{}) string {
	if tmp, ok := v.(iClassName); ok {
		return tmp.ClassName()
	} else {
		t := reflect.TypeOf(v)
		return t.Elem().Name()
	}
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
		cname := getClassName(v)
		p = path.Join("classes", cname)
	}

	p = path.Join(ParseVersion, p)
	return p
}

type Config map[string]interface{}

type configRequestT struct{}

func (c *configRequestT) method() string {
	return "GET"
}

func (c *configRequestT) endpoint() (string, error) {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = parseHost
	u.Path = path.Join(ParseVersion, "config")
	return u.String(), nil
}

func (c *configRequestT) body() (string, error) {
	return "", nil
}

func (c *configRequestT) useMasterKey() bool {
	return false
}

func (c *configRequestT) session() *sessionT {
	return nil
}

func (c *configRequestT) contentType() string {
	return ""
}

func GetConfig() (Config, error) {
	b, err := defaultClient.doRequest(&configRequestT{})
	if err != nil {
		return nil, err
	}

	c := struct {
		Params Config `json:"params"`
	}{}
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}

	return c.Params, nil
}
