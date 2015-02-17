package parse

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestLogin(t *testing.T) {
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.Form.Get("username") != "username" {
			t.Errorf("login request did not include proper username. got [%v] expected [%v]\n", r.Form.Get("username"), "username")
		}

		if r.Form.Get("password") != "password" {
			t.Errorf("login request did not include proper password. got [%v] expected [%v]\n", r.Form.Get("password"), "password")
		}

		fmt.Fprintf(w, `{"sessionToken":"abcd","username":"kylemcc@gmail.com","createdAt":"2014-04-01T14:44:14.123Z","updatedAt":"2014-12-01T12:34:56.789Z"}`)
	})
	defer teardownTestServer()

	s, err := Login("username", "password", nil)
	if err != nil {
		t.Errorf("unexpected error on login: %v\n", err)
		t.FailNow()
	}

	st := s.(*sessionT)
	if st.sessionToken != "abcd" {
		t.Errorf("login did not set a proper session token. got: [%v] expected: [%v]\n", st.sessionToken, "abcd")
	}

	u := s.User()
	expectedUser := &User{
		Username: "kylemcc@gmail.com",
		Base: Base{
			CreatedAt: time.Date(2014, 4, 1, 14, 44, 14, 123000000, time.UTC),
			UpdatedAt: time.Date(2014, 12, 1, 12, 34, 56, 789000000, time.UTC),
			Extra: map[string]interface{}{
				"SessionToken": "abcd",
			},
		},
	}

	if !reflect.DeepEqual(u, expectedUser) {
		t.Errorf("login did not return correct user. Got:\n[%+v]\nexpected:\n[%+v]\n", u, expectedUser)
	}
}

type CustomUser struct {
	User
	Phone string
	City  string
}

func TestLoginCustomUserType(t *testing.T) {
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.Form.Get("username") != "username" {
			t.Errorf("login request did not include proper username. got [%v] expected [%v]\n", r.Form.Get("username"), "username")
		}

		if r.Form.Get("password") != "password" {
			t.Errorf("login request did not include proper password. got [%v] expected [%v]\n", r.Form.Get("password"), "password")
		}

		fmt.Fprintf(w, `{"sessionToken":"abcd","username":"kylemcc@gmail.com","createdAt":"2014-04-01T14:44:14.123Z","updatedAt":"2014-12-01T12:34:56.789Z","phone":"3105551234","city":"Santa Monica"}`)
	})
	defer teardownTestServer()

	s, err := Login("username", "password", &CustomUser{})
	if err != nil {
		t.Errorf("unexpected error on login: %v\n", err)
		t.FailNow()
	}

	st := s.(*sessionT)
	if st.sessionToken != "abcd" {
		t.Errorf("login did not set a proper session token. got: [%v] expected: [%v]\n", st.sessionToken, "abcd")
	}

	u := s.User()
	expectedUser := &CustomUser{
		User: User{
			Username: "kylemcc@gmail.com",
			Base: Base{
				CreatedAt: time.Date(2014, 4, 1, 14, 44, 14, 123000000, time.UTC),
				UpdatedAt: time.Date(2014, 12, 1, 12, 34, 56, 789000000, time.UTC),
				Extra: map[string]interface{}{
					"SessionToken": "abcd",
				},
			},
		},
		Phone: "3105551234",
		City:  "Santa Monica",
	}

	if !reflect.DeepEqual(u, expectedUser) {
		t.Errorf("login did not return correct user. Got:\n[%+v]\nexpected:\n[%+v]\n", u, expectedUser)
	}
}

func TestSessionOperationsSetSessionTokenHeader(t *testing.T) {
	opCreate := 1
	opDelete := 2
	opUpdate := 3
	opQuery := 4

	var currentOp int
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get(AppIdHeader); h != "app_id" {
			t.Errorf("request did not have App ID header set!")
		}

		if h := r.Header.Get(SessionTokenHeader); h != "session_token" {
			t.Errorf("request did not have Session Token header set!")
		}

		if h := r.Header.Get(RestKeyHeader); h != "rest_key" {
			t.Errorf("request did not have Rest Key header set!")
		}

		if h := r.Header.Get(MasterKeyHeader); h != "" {
			t.Errorf("request had Master Key header set!")
		}

		switch currentOp {
		case opQuery:
			fmt.Fprintf(w, `{"results":[{}]}`)
		default:
			fmt.Fprintf(w, `{}`)
		}
	})

	var s Session
	s = &sessionT{
		user:         &User{},
		sessionToken: "session_token",
	}

	currentOp = opCreate
	if err := s.Create(&User{}); err != nil {
		t.Errorf("unexpected error on Session.Create: %v\n", err)
	}

	currentOp = opDelete
	if err := s.Delete(&User{}); err != nil {
		t.Errorf("unexpected error on Session.Delete: %v\n", err)
	}

	u, err := s.NewUpdate(&User{})
	if err != nil {
		t.Errorf("unexpected error on Session.NewUpdate: %v\n", err)
	}
	u.Set("key", "value")
	currentOp = opUpdate
	if err := u.Execute(); err != nil {
		t.Errorf("unexpected error executing update: %v\n", err)
	}

	q, err := s.NewQuery(&User{})
	if err != nil {
		t.Errorf("unexpected error on Session.NewQuery: %v\n", err)
	}
	q.EqualTo("key", "value")
	currentOp = opQuery
	if err := q.First(); err != nil {
		t.Errorf("unexpected error executing query: %v\n", err)
	}
}
