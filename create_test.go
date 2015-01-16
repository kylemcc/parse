package parse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestCreateRequiresPointer(t *testing.T) {
	u := User{}
	expected := "v must be a non-nil pointer"
	if err := Create(u, false); err == nil {
		t.Error("Create should return an error when argument is not a pointer")
	} else if err.Error() != expected {
		t.Errorf("Unexpected error message. Got [%s] expected [%s]\n", err, expected)
	}
}

type TestUser struct {
	FirstName string
	LastName  string
	Email     string
	Ignore    string `parse:"-"`
	FCount    int    `parse:"followers"`
	Base
}

func TestPayload(t *testing.T) {
	tu := TestUser{
		FirstName: "Kyle",
		LastName:  "M",
		Email:     "kylemcc@gmail.com",
		Ignore:    "shouldn't appear in payload",
		FCount:    11,
	}

	cr := createT{
		v: &tu,
	}

	e := map[string]interface{}{
		"firstName": "Kyle",
		"lastName":  "M",
		"email":     "kylemcc@gmail.com",
		"followers": 11,
	}

	expected := map[string]interface{}{}
	eb, _ := json.Marshal(e)
	_ = json.Unmarshal(eb, &expected)

	actual := map[string]interface{}{}
	b, err := cr.body()
	if err != nil {
		t.Errorf("unexpected error generating payload: %v\n", err)
		t.FailNow()
	}

	err = json.Unmarshal([]byte(b), &actual)
	if err != nil {
		t.Errorf("unexpected error unmarshaling payload: %v\n", err)
		t.FailNow()

	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("payload different from expected. expected:\n%s\n\ngot:\n%s\n", eb, b)
	}
}

func TestCreate(t *testing.T) {
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get(AppIdHeader); h != "app_id" {
			t.Errorf("request did not have App ID header set!")
		}

		if h := r.Header.Get(RestKeyHeader); h != "rest_key" {
			t.Errorf("request did not have Rest Key header set!")
		}

		if h := r.Header.Get(SessionTokenHeader); h != "" {
			t.Errorf("request had Session Token header set!")
		}

		if h := r.Header.Get(MasterKeyHeader); h != "" {
			t.Errorf("request had Master Key header set!")
		}

		fmt.Fprintf(w, `{"createdAt":"2014-12-19T18:05:57Z","objectId":"abcDEF"}`)
	})
	defer teardownTestServer()

	u := TestUser{
		FirstName: "Kyle",
		LastName:  "M",
		Email:     "kylemcc@gmail.com",
		FCount:    11,
	}

	err := Create(&u, false)
	if err != nil {
		t.Errorf("Unexpected error creating object: %v\n", err)
		t.FailNow()
	}

	if u.Id != "abcDEF" {
		t.Errorf("Create did not set proper id on instance. u.Id: %v\n", u.Id)
	}

	if u.CreatedAt != time.Date(2014, 12, 19, 18, 5, 57, 0, time.UTC) {
		t.Errorf("Create did not set proper createdAt date. u.CreatedAt: %v\n", u.CreatedAt)
	}
}

func TestCreateUseMasterKey(t *testing.T) {
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get(AppIdHeader); h != "app_id" {
			t.Errorf("request did not have App ID header set!")
		}

		if h := r.Header.Get(RestKeyHeader); h != "" {
			t.Errorf("request had Rest Key header set!")
		}

		if h := r.Header.Get(SessionTokenHeader); h != "" {
			t.Errorf("request had Session Token header set!")
		}

		if h := r.Header.Get(MasterKeyHeader); h != "master_key" {
			t.Errorf("request did not have Master Key header set!")
		}

		fmt.Fprintf(w, `{"createdAt":"2014-12-19T18:05:57Z","objectId":"abcDEF"}`)
	})
	defer teardownTestServer()

	u := TestUser{
		FirstName: "Kyle",
		LastName:  "M",
		Email:     "kylemcc@gmail.com",
		FCount:    11,
	}

	err := Create(&u, true)
	if err != nil {
		t.Errorf("unexpected error on create: %v\n", err)
	}
}

type TestTypeOmitEmpty struct {
	StrField   string
	OEStrField string `parse:",omitempty"`

	BoolField   bool
	OEBoolField bool `parse:",omitempty"`

	IntField   int
	OEIntField int `parse:",omitempty"`

	ArrField   []string
	OEArrField []string `parse:",omitempty"`

	EmptyArrField   []string
	OEEmptyArrField []string `parse:",omitempty"`

	MapField   map[string]interface{}
	OEMapField map[string]interface{} `parse:",omitempty"`

	EmptyMapField   map[string]interface{}
	OEEmptyMapField map[string]interface{} `parse:",omitempty"`

	PtrField   *User
	OEPtrField *User `parse:",omitempty"`
}

func TestCreateOmitEmpty(t *testing.T) {
	cr := &createT{
		v: &TestTypeOmitEmpty{
			EmptyArrField:   []string{},
			OEEmptyArrField: []string{},
			EmptyMapField:   map[string]interface{}{},
			OEEmptyMapField: map[string]interface{}{},
		},
	}

	e := map[string]interface{}{
		"strField":      "",
		"boolField":     false,
		"intField":      0,
		"arrField":      nil,
		"emptyArrField": []string{},
		"mapField":      nil,
		"emptyMapField": map[string]interface{}{},
		"ptrField":      nil,
	}

	expected := map[string]interface{}{}
	eb, _ := json.Marshal(e)
	_ = json.Unmarshal(eb, &expected)

	actual := map[string]interface{}{}
	b, err := cr.body()
	if err != nil {
		t.Errorf("unexpected error generating payload: %v\n", err)
		t.FailNow()
	}

	err = json.Unmarshal([]byte(b), &actual)
	if err != nil {
		t.Errorf("unexpected error unmarshaling payload: %v\n", err)
		t.FailNow()

	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("payload different from expected. expected:\n%s\n\ngot:\n%s\n", eb, b)
	}
}
