package parse

import "net/url"

type Session interface {
	User() *User
	NewQuery(v interface{}) (Query, error)
	NewUpdate(v interface{}) (Update, error)
	Create(v interface{}) error
	Delete(v interface{}) error
}

type loginRequestT struct {
	username string
	password string
}

type sessionT struct {
	user         *User
	sessionToken string
}

func Login(username, password string) (Session, error) {
	s := &sessionT{user: &User{}}
	err := defaultClient.doRequest(&loginRequestT{username: username, password: password}, s.user)
	if err != nil {
		return nil, err
	}

	if st, ok := s.user.Extra["SessionToken"]; ok {
		if stStr, ok := st.(string); ok {
			s.sessionToken = stStr
		}
	}
	return s, nil
}

func (s *sessionT) User() *User {
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

func (s *loginRequestT) method() string {
	return "GET"
}

func (s *loginRequestT) endpoint() (string, error) {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = parseHost
	u.Path = "/1/login"

	v := url.Values{}
	v["username"] = []string{s.username}
	v["password"] = []string{s.password}
	u.RawQuery = v.Encode()

	return u.String(), nil
}

func (s *loginRequestT) body() (string, error) {
	return "", nil
}

func (s *loginRequestT) useMasterKey() bool {
	return false
}

func (s *loginRequestT) session() *sessionT {
	return nil
}

func (s *loginRequestT) contentType() string {
	return "application/x-www-form-urlencoded"
}
