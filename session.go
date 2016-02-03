package parse

import (
	"encoding/json"
	"errors"
	"net/url"
	"reflect"
)

type Session interface {
	User() interface{}
	NewQuery(v interface{}) (Query, error)
	NewUpdate(v interface{}) (Update, error)
	Create(v interface{}) error
	Delete(v interface{}) error
	CallFunction(name string, params Params, resp interface{}) error
}

type loginRequestT struct {
	username string
	password string
	s        *sessionT
	authdata *AuthData
}

type sessionT struct {
	user         interface{}
	sessionToken string
}

// Login in as the user identified by the provided username and password.
//
// Optionally provide a custom User type to use in place of parse.User. If u is not
// nil, it will be populated with the user's attributes, and will be accessible
// by calling session.User().
func Login(username, password string, u interface{}) (Session, error) {
	var user interface{}

	if u == nil {
		user = &User{}
	} else if err := validateUser(u); err != nil {
		return nil, err
	} else {
		user = u
	}

	s := &sessionT{user: user}
	if b, err := defaultClient.doRequest(&loginRequestT{username: username, password: password}); err != nil {
		return nil, err
	} else if st, err := handleLoginResponse(b, s.user); err != nil {
		return nil, err
	} else {
		s.sessionToken = st
	}

	return s, nil
}

func LoginFacebook(authData *FacebookAuthData, u interface{}) (Session, error) {
	var user interface{}

	if u == nil {
		user = &User{}
	} else if err := validateUser(u); err != nil {
		return nil, err
	} else {
		user = u
	}

	s := &sessionT{user: user}
	if b, err := defaultClient.doRequest(&loginRequestT{authdata: &AuthData{Facebook: authData}}); err != nil {
		return nil, err
	} else if st, err := handleLoginResponse(b, s.user); err != nil {
		return nil, err
	} else {
		s.sessionToken = st
	}

	return s, nil
}

// Log in as the user identified by the session token st
//
// Optionally provide a custom User type to use in place of parse.User. If user is
// not nil, it will be populated with the user's attributes, and will be accessible
// by calling session.User().
func Become(st string, u interface{}) (Session, error) {
	var user interface{}

	if u == nil {
		user = &User{}
	} else if err := validateUser(u); err != nil {
		return nil, err
	} else {
		user = u
	}

	r := &loginRequestT{
		s: &sessionT{
			sessionToken: st,
			user:         user,
		},
	}

	if b, err := defaultClient.doRequest(r); err != nil {
		return nil, err
	} else if err := handleResponse(b, r.s.user); err != nil {
		return nil, err
	}
	return r.s, nil
}

func (s *sessionT) User() interface{} {
	return s.user
}

func (s *sessionT) NewQuery(v interface{}) (Query, error) {
	q, err := NewQuery(v)
	if err == nil {
		if qt, ok := q.(*queryT); ok {
			qt.currentSession = s
		}
	}
	return q, err
}

func (s *sessionT) NewUpdate(v interface{}) (Update, error) {
	u, err := NewUpdate(v)
	if err == nil {
		if ut, ok := u.(*updateT); ok {
			ut.currentSession = s
		}
	}
	return u, err
}

func (s *sessionT) Create(v interface{}) error {
	return create(v, false, s)
}

func (s *sessionT) Delete(v interface{}) error {
	return _delete(v, false, s)
}

func (s *sessionT) CallFunction(name string, params Params, resp interface{}) error {
	return callFn(name, params, resp, s)
}

func (s *loginRequestT) method() string {
	if s.authdata != nil {
		return "POST"
	}

	return "GET"
}

func (s *loginRequestT) endpoint() (string, error) {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = parseHost
	if s.s != nil {
		u.Path = "/1/users/me"
	} else if s.authdata != nil {
		u.Path = "/1/users"
	} else {
		u.Path = "/1/login"
	}

	if s.username != "" && s.password != "" {
		v := url.Values{}
		v["username"] = []string{s.username}
		v["password"] = []string{s.password}
		u.RawQuery = v.Encode()
	}

	return u.String(), nil
}

func (s *loginRequestT) body() (string, error) {
	if s.authdata != nil {
		b, err := json.Marshal(map[string]interface{}{"authData": s.authdata})
		return string(b), err
	}
	return "", nil
}

func (s *loginRequestT) useMasterKey() bool {
	return false
}

func (s *loginRequestT) session() *sessionT {
	return s.s
}

func (s *loginRequestT) contentType() string {
	return "application/x-www-form-urlencoded"
}

func validateUser(u interface{}) error {
	rv := reflect.ValueOf(u)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("u must be a non-nil pointer")
	} else if getClassName(u) != "_User" {
		return errors.New("u must embed parse.User or implement a ClassName function that returns \"_User\"")
	}
	return nil
}

func handleLoginResponse(body []byte, dst interface{}) (sessionToken string, err error) {
	data := make(map[string]interface{})
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	st, ok := data["sessionToken"]
	if !ok {
		return "", errors.New("response did not contain sessionToken")
	}
	return st.(string), populateValue(dst, data)
}
