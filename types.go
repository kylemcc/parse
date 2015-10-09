package parse

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"path"
	"reflect"
	"time"
)

func init() {
	gob.Register(&aclT{})
}

var registeredTypes = map[string]reflect.Type{}

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

type AnonymousAuthData struct {
	Id string `json:"id"`
}

type TwitterAuthData struct {
	Id              string `json:"id"`
	ScreenName      string `json:"screen_name" parse:"screen_name"`
	ConsumerKey     string `json:"consumer_key" parse:"consumer_key"`
	ConsumerSecret  string `json:"consumer_secret" parse:"consumer_secret"`
	AuthToken       string `json:"auth_token" parse:"auth_token"`
	AuthTokenSecret string `json:"auth_token_secret" parse:"auth_token_secret"`
}

type FacebookAuthData struct {
	Id             string
	AccessToken    string    `parse:"access_token"`
	ExpirationDate time.Time `parse:"expiration_date"`
}

func (a *FacebookAuthData) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id             string `json:"id"`
		AccessToken    string `json:"access_token" parse:"access_token"`
		ExpirationDate string `json:"expiration_date"`
	}{
		a.Id, a.AccessToken, a.ExpirationDate.Format("2006-01-02T15:04:05.000Z"),
	})
}

func (a *FacebookAuthData) UnmarshalJSON(b []byte) (err error) {
	data := struct {
		Id             string `json:"id"`
		AccessToken    string `json:"access_token" parse:"access_token"`
		ExpirationDate string `json:"expiration_date"`
	}{}

	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	a.Id = data.Id
	a.AccessToken = data.AccessToken
	a.ExpirationDate, err = time.Parse("2006-01-02T15:04:05.000Z", data.ExpirationDate)
	return err
}

type AuthData struct {
	Twitter   *TwitterAuthData   `json:"twitter,omitempty"`
	Facebook  *FacebookAuthData  `json:"facebook,omitempty"`
	Anonymous *AnonymousAuthData `json:"anonymous,omitempty"`
}

// Represents the built-in Parse "User" class. Embed this type in a custom
// type containing any custom fields. When fetching user objects, any retrieved
// fields with no matching struct field will be stored in User.Extra (map[string]interface{})
type User struct {
	Base
	Username      string
	Email         string
	EmailVerified bool `json:"-" parse:"-"`
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
	Badge          int      `parse:",omitempty"`
	Channels       []string `parse:",omitempty"`
	TimeZone       string
	DeviceType     string
	PushType       string `parse:",omitempty"`
	GCMSenderId    string `parse:",omitempty"`
	InstallationId string
	DeviceToken    string   `parse:",omitempty"`
	ChannelUris    []string `parse:",omitempty"`
	AppName        string
	AppVersion     string
	ParseVersion   string
	AppIdentifier  string
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

func (a *aclT) GobEncode() ([]byte, error) {
	return json.Marshal(a)
}

func (a *aclT) GobDecode(b []byte) error {
	return json.Unmarshal(b, &a)
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

// Returns this distance from this GeoPoint to another in radians
func (g GeoPoint) RadiansTo(point GeoPoint) float64 {
	d2r := math.Pi / 180.0
	lat1Rad := g.Latitude * d2r
	long1Rad := g.Longitude * d2r
	lat2Rad := point.Latitude * d2r
	long2Rad := point.Longitude * d2r

	sinDeltaLatDiv2 := math.Sin((lat1Rad - lat2Rad) / 2)
	sinDeltaLongDiv2 := math.Sin((long1Rad - long2Rad) / 2)

	// Square of half the straight line chord distance between both points.
	var a = sinDeltaLatDiv2*sinDeltaLatDiv2 + math.Cos(lat1Rad)*math.Cos(lat2Rad)*sinDeltaLongDiv2*sinDeltaLongDiv2
	a = math.Min(1.0, a)
	return 2 * math.Asin(math.Sqrt(a))
}

// Returns this distance from this GeoPoint to another in kilometers
func (g GeoPoint) KilometersTo(point GeoPoint) float64 {
	return g.RadiansTo(point) * 6371.0
}

// Returns this distance from this GeoPoint to another in miles
func (g GeoPoint) MilesTo(point GeoPoint) float64 {
	return g.RadiansTo(point) * 3958.8
}

// Represents the Parse File type
type File struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func (f *File) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name string `json:"name"`
		Url  string `json:"url"`
		Type string `json:"__type"`
	}{
		f.Name, f.Url, "File",
	})
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

	t, err := time.Parse("2006-01-02T15:04:05.000Z", s.Iso)
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
		cname := getClassName(inst)
		p = path.Join("classes", cname)
	}

	p = path.Join(ParseVersion, p)
	return p
}

type Config map[string]interface{}

// Retrieves the value associated with the given key, and,
// if present, converts the value to a string and returns
// it. If the value is not present, or is not a string
// value, an empty string is returned
func (c Config) String(key string) string {
	if v, ok := c[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Retrieves the value associated with the given key, and,
// if present, converts the value to a byte slice and returns
// it. If the value is not present, or is not a string
// value, an empty byte slice is returned
func (c Config) Bytes(key string) []byte {
	if v, ok := c[key]; ok {
		if s, ok := v.(string); ok {
			return []byte(s)
		}
	}
	return make([]byte, 0, 0)
}

// Retrieves the value associated with the given key, and,
// if present, converts the value to a bool and returns
// it. If the value is not present, or is not a bool
// value, false is returned
func (c Config) Bool(key string) bool {
	if v, ok := c[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// Retrieves the value associated with the given key, and,
// if present, converts the value to an int and returns
// it. If the value is not present, or is not a numeric
// value, 0 is returned
func (c Config) Int(key string) int {
	if v, ok := c[key]; ok {
		// since we're unmarshaling into an interface{} value, all
		// numbers will be float64 values
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return 0
}

// Retrieves the value associated with the given key, and,
// if present, converts the value to an int64 and returns
// it. If the value is not present, or is not a numeric
// value, 0 is returned
func (c Config) Int64(key string) int64 {
	if v, ok := c[key]; ok {
		// since we're unmarshaling into an interface{} value, all
		// numbers will be float64 values
		if f, ok := v.(float64); ok {
			return int64(f)
		}
	}
	return 0
}

// Retrieves the value associated with the given key, and,
// if present, converts the value to an float64 and returns
// it. If the value is not present, or is not a numeric
// value, 0 is returned
func (c Config) Float(key string) float64 {
	if v, ok := c[key]; ok {
		// since we're unmarshaling into an interface{} value, all
		// numbers will be float64 values
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

// Retrieves the value associated with the given key, and,
// if present, converts the value to a slice of interface{}
// values and returns it. If the value is not present, or
// is not an array value, nil is returned
func (c Config) Values(key string) []interface{} {
	if v, ok := c[key]; ok {
		if s, ok := v.([]interface{}); ok {
			return s
		}
	}
	return nil
}

// Retrieves the value associated with the given key, and,
// if present, converts the value to a slice of string
// values and returns it. If the value is not present, or
// is not an array value, nil is returned
func (c Config) Strings(key string) []string {
	if v, ok := c[key]; ok {
		if is, ok := v.([]interface{}); ok {
			ss := []string{}
			for _, i := range is {
				if s, ok := i.(string); ok {
					ss = append(ss, s)
				}
			}
			if len(ss) == len(is) {
				return ss
			}
		}
	}
	return nil
}

// Retrieves the value associated with the given key, and,
// if present, converts the value to a slice of int values
// and returns it. If the value is not present, or is not
// an array value, nil is returned
func (c Config) Ints(key string) []int {
	if v, ok := c[key]; ok {
		if ifs, ok := v.([]interface{}); ok {
			ints := []int{}
			for _, i := range ifs {
				if f, ok := i.(float64); ok {
					ints = append(ints, int(f))
				}
			}
			if len(ints) == len(ifs) {
				return ints
			}
		}
	}
	return nil
}

// Retrieves the value associated with the given key, and,
// if present, converts the value to a slice of int64 values
// and returns it. If the value is not present, or is not
// an array value, nil is returned
func (c Config) Int64s(key string) []int64 {
	if v, ok := c[key]; ok {
		if ifs, ok := v.([]interface{}); ok {
			ints := []int64{}
			for _, i := range ifs {
				if f, ok := i.(float64); ok {
					ints = append(ints, int64(f))
				}
			}
			if len(ints) == len(ifs) {
				return ints
			}
		}
	}
	return nil
}

// Retrieves the value associated with the given key, and,
// if present, converts the value to a slice of float64 values
// and returns it. If the value is not present, or is not
// an array value, nil is returned
func (c Config) Floats(key string) []float64 {
	if v, ok := c[key]; ok {
		if is, ok := v.([]interface{}); ok {
			fs := []float64{}
			for _, i := range is {
				if f, ok := i.(float64); ok {
					fs = append(fs, f)
				}
			}
			if len(fs) == len(is) {
				return fs
			}
		}
	}
	return nil
}

// Retrieves the value associated with the given key, and, if present,
// converts value to a Config type (map[string]interface{}) and returns
// it. If the value is not present, or is not a JSON object, nil is
// returned
func (c Config) Map(key string) Config {
	if v, ok := c[key]; ok {
		if s, ok := v.(map[string]interface{}); ok {
			return Config(s)
		}
	}
	return nil
}

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

// Register a type so that it can be handled when populating struct values.
//
// The provided value will be registered under the name provided by the ClassName method
// if it is implemented, otherwise by the name of the type. When handling Parse responses,
// any object value with __type "Object" or "Pointer" and className matching the type provided
// will be unmarshaled into pointer to the provided type.
//
// This is useful in at least one instance: If you have an array or object field on a
// Parse class that contains pointers to or instances of Objects of arbitrary types
// that cannot be represented by a single type on your struct definition, but you would
// still like to be able to populate your struct with these values.
//
// In order to accomplish this, the field in question on your struct definition
// should either be of type interface{}, or another interface type that all possible
// types implement.
//
// Accepts a value t, representing the type to be registered. The value
// t should be either a struct value, or a pointer to a struct. Otherwise,
// an error will be returned.
func RegisterType(t interface{}) error {
	rv := reflect.ValueOf(t)
	rvi := reflect.Indirect(rv)

	if rvi.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct or pointer to struct, got: %v", rv.Kind())
	}

	className := getClassName(t)
	registeredTypes[className] = rvi.Type()
	return nil
}

// Transform the given value into the proper representation for Marshaling as part
// of a request
//
// E.g. A struct is turned into a Pointer type, a time.Time is turned into a Date
func encodeForRequest(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	rv := reflect.ValueOf(v)
	rvi := reflect.Indirect(rv)
	rt := rvi.Type()
	if rt.Kind() == reflect.Struct {
		switch v.(type) {
		case time.Time, *time.Time, Date, *Date:
			switch v.(type) {
			case time.Time:
				return Date(v.(time.Time))
			case *time.Time:
				return Date(*v.(*time.Time))
			case Date, *Date:
				return v
			case string:
				return map[string]interface{}{
					"__type": "Date",
					"iso":    v.(string),
				}
			}
		case Pointer, *Pointer:
			return v
		case GeoPoint, *GeoPoint:
			return v
		case ACL, *ACL:
			return v
		case AuthData, *AuthData:
			b, _ := json.Marshal(v)
			return string(b)
		default:
			var cname string

			if tmp, ok := reflect.Zero(rvi.Type()).Interface().(iClassName); ok {
				cname = tmp.ClassName()
			} else if tmp, ok := reflect.New(rvi.Type()).Interface().(iClassName); ok {
				cname = tmp.ClassName()
			} else {
				cname = rt.Name()
			}

			if idf := rvi.FieldByName("Id"); idf.IsValid() {
				id := idf.Interface().(string)
				return Pointer{
					Id:        id,
					ClassName: cname,
				}
			}
		}
	} else if rt.Kind() == reflect.Slice {
		vals := make([]interface{}, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			vals = append(vals, encodeForRequest(rv.Index(i).Interface()))
		}
		return vals
	}

	return v
}
