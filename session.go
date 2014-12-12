package parse

type Session interface {
	User() *User
	NewQuery(v interface{}) (Query, error)
}

type sessionT struct {
	user         *User
	sessionToken string
}

func Login(username, password string) (Session, error) {
	return nil, nil
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

func (s *sessionT) User() *User {
	return s.user
}
