package parse

import (
	"encoding/json"
	"reflect"
	"testing"
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
