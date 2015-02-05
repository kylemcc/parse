package parse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestUpdateRequiresPointer(t *testing.T) {
	u := User{}
	expected := "v must be a non-nil pointer"
	if _, err := NewUpdate(u); err == nil {
		t.Error("NewUpdate should return an error when argument is not a pointer")
	} else if err.Error() != expected {
		t.Errorf("Unexpected error message. Got [%s] expected [%s]\n", err, expected)
	}
}

func TestOperations(t *testing.T) {
	type UpdateTest struct {
		F1  string
		F2  time.Time
		F3  time.Time
		F4  *User
		F5  User
		F6  bool `parse:"f5_custom"`
		F7  int
		F8  uint
		F9  float32
		F10 string
		F11 []string
		F12 []string
		F13 []string
	}

	u, err := NewUpdate(&UpdateTest{})
	if err != nil {
		t.Errorf("Unexpected error creating update: %v\n", err)
		t.FailNow()
	}

	u.Set("f1", "string")
	u.Set("f3", time.Date(2014, 12, 20, 18, 31, 19, 123000000, time.UTC))
	u.Set("f4", User{Base: Base{Id: "abcd"}})
	u.Set("f6_custom", true)
	u.Increment("f7", 1)
	u.Increment("f8", 2)
	u.Increment("f9", 3.2)
	u.Delete("f10")
	u.Add("f11", "abc", "def")
	u.AddUnique("f12", "123", "456")
	u.Remove("f13", "zyx", "wvu")

	acl := NewACL()
	acl.SetPublicReadAccess(true)
	acl.SetWriteAccess("abc", true)
	acl.SetRoleWriteAccess("xyz", true)

	u.SetACL(acl)

	em := map[string]interface{}{
		"f1": "string",
		"f3": map[string]interface{}{
			"__type": "Date",
			"iso":    "2014-12-20T18:31:19.123Z",
		},
		"f4": map[string]interface{}{
			"__type":    "Pointer",
			"className": "_User",
			"objectId":  "abcd",
		},
		"f6_custom": true,
		"f7": map[string]interface{}{
			"__op":   "Increment",
			"amount": 1,
		},
		"f8": map[string]interface{}{
			"__op":   "Increment",
			"amount": 2,
		},
		"f9": map[string]interface{}{
			"__op":   "Increment",
			"amount": 3.2,
		},
		"f10": map[string]interface{}{
			"__op": "Delete",
		},
		"f11": map[string]interface{}{
			"__op": "Add",
			"objects": []interface{}{
				"abc", "def",
			},
		},
		"f12": map[string]interface{}{
			"__op": "AddUnique",
			"objects": []interface{}{
				"123", "456",
			},
		},
		"f13": map[string]interface{}{
			"__op": "Remove",
			"objects": []interface{}{
				"zyx", "wvu",
			},
		},
		"ACL": map[string]map[string]bool{
			"*": {
				"read": true,
			},
			"abc": {
				"write": true,
			},
			"role:xyz": {
				"write": true,
			},
		},
	}

	expected := map[string]interface{}{}
	eb, _ := json.Marshal(em)
	_ = json.Unmarshal(eb, &expected)

	b, err := u.(*updateT).body()
	if err != nil {
		t.Errorf("error marshaling where: %v\n", err)
		t.FailNow()
	}

	actual := map[string]interface{}{}
	err = json.Unmarshal([]byte(b), &actual)
	if err != nil {
		t.Errorf("error unmarshaling update: %v\n", err)
		t.FailNow()
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("update different from expected. expected:\n%s\n\ngot:\n%s\n", eb, b)
	}
}

func TestExecuteUpdatesStruct(t *testing.T) {
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"updatedAt":"2014-12-20T18:23:49.123Z","f11":["abc","abc","def"],"f12":["123","456"],"f13":["tsr"]}`)
	})
	defer teardownTestServer()

	type UpdateTest struct {
		F1        string
		F2        time.Time
		F3        time.Time
		F4        User
		F5        User
		F6        bool `parse:"f6_custom"`
		F7        int
		F8        uint
		F9        float32
		F10       string
		F11       []string
		F12       []string
		F13       []string
		Id        string
		UpdatedAt time.Time
	}

	tu := UpdateTest{
		F1:  "not string",
		F2:  time.Now(),
		F3:  time.Now(),
		F7:  10,
		F8:  73,
		F9:  4.8,
		F10: "not empty",
		F11: []string{"abc"},
		F12: []string{"123"},
		F13: []string{"zyx", "wvu", "tsr"},
	}

	u, err := NewUpdate(&tu)
	if err != nil {
		t.Errorf("Unexpected error creating update: %v\n", err)
		t.FailNow()
	}

	u.Set("f1", "string")
	u.Set("f2", time.Date(2014, 12, 20, 18, 31, 19, 123000000, time.UTC))
	u.Set("f3", time.Date(2014, 12, 20, 18, 31, 19, 123000000, time.UTC))
	u.Set("f4", User{Base: Base{Id: "abcd"}})
	u.Set("f5", User{Base: Base{Id: "efghi"}})
	u.Set("f6_custom", true)
	u.Increment("f7", 1)
	u.Increment("f8", 2)
	u.Increment("f9", 3.2)
	u.Delete("f10")
	u.Add("f11", "abc", "def")
	u.AddUnique("f12", "123", "456")
	u.Remove("f13", "zyx", "wvu")

	if err := u.Execute(); err != nil {
		t.Errorf("Unexpected error executing update: %v\n", err)
		t.FailNow()
	}

	tuExpected := UpdateTest{
		F1:        "string",
		F2:        time.Date(2014, 12, 20, 18, 31, 19, 123000000, time.UTC),
		F3:        time.Date(2014, 12, 20, 18, 31, 19, 123000000, time.UTC),
		F4:        User{Base: Base{Id: "abcd"}},
		F5:        User{Base: Base{Id: "efghi"}},
		F6:        true,
		F7:        11,
		F8:        75,
		F9:        8.0,
		F10:       "",
		F11:       []string{"abc", "abc", "def"},
		F12:       []string{"123", "456"},
		F13:       []string{"tsr"},
		UpdatedAt: time.Date(2014, 12, 20, 18, 23, 49, 123000000, time.UTC),
	}

	if !reflect.DeepEqual(tu, tuExpected) {
		//if tu != tuExpected {
		t.Errorf("Update did not properly modify struct. Got:\n[%+v]\nexpected:\n[%+v]\n", tu, tuExpected)
	}
}

func TestUpdateUseMasterKey(t *testing.T) {
	shouldHaveMasterKey := false
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get(AppIdHeader); h != "app_id" {
			t.Errorf("request did not have App ID header set!")
		}

		if h := r.Header.Get(SessionTokenHeader); h != "" {
			t.Errorf("request had Session Token header set!")
		}

		if shouldHaveMasterKey {
			if h := r.Header.Get(RestKeyHeader); h != "" {
				t.Errorf("request had Rest Key header set!")
			}

			if h := r.Header.Get(MasterKeyHeader); h != "master_key" {
				t.Errorf("request did not have Master Key header set!")
			}
		} else {
			if h := r.Header.Get(RestKeyHeader); h != "rest_key" {
				t.Errorf("request did not have Rest Key header set!")
			}

			if h := r.Header.Get(MasterKeyHeader); h != "" {
				t.Errorf("request had Master Key header set!")
			}
		}

		fmt.Fprintf(w, `{"updatedAt":"2014-12-20T18:23:49.123Z","f11":["abc","abc","def"],"f12":["123","456"],"f13":["tsr"]}`)
	})
	defer teardownTestServer()

	u1, _ := NewUpdate(&User{})
	u1.Set("city", "Chicago")
	if err := u1.Execute(); err != nil {
		t.Errorf("Unexpected error executing update: %v\n", err)
	}

	u2, _ := NewUpdate(&User{})
	u2.Set("city", "Chicago")
	u2.UseMasterKey()
	shouldHaveMasterKey = true
	if err := u2.Execute(); err != nil {
		t.Errorf("Unexpected error executing update: %v\n", err)
	}
}
