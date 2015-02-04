package parse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestACLMarshal(t *testing.T) {
	acl := NewACL()

	acl.SetPublicReadAccess(true)

	acl.SetReadAccess("abc", true)
	acl.SetReadAccess("def", false)
	acl.SetReadAccess("ghi", true)

	acl.SetWriteAccess("def", false)
	acl.SetWriteAccess("ghi", true)
	acl.SetWriteAccess("jkl", true)

	acl.SetRoleReadAccess("zyx", true)
	acl.SetRoleReadAccess("wvu", false)
	acl.SetRoleReadAccess("tsr", true)

	acl.SetRoleWriteAccess("wvu", false)
	acl.SetRoleWriteAccess("tsr", true)
	acl.SetRoleWriteAccess("qpo", true)

	expected := map[string]map[string]bool{
		"*": map[string]bool{
			"read": true,
		},
		"abc": map[string]bool{
			"read": true,
		},
		"ghi": map[string]bool{
			"read":  true,
			"write": true,
		},
		"jkl": map[string]bool{
			"write": true,
		},
		"role:zyx": map[string]bool{
			"read": true,
		},
		"role:tsr": map[string]bool{
			"read":  true,
			"write": true,
		},
		"role:qpo": map[string]bool{
			"write": true,
		},
	}

	b, err := json.Marshal(acl)
	if err != nil {
		t.Errorf("unexpected error marshaling ACL: %v\n", err)
		t.FailNow()
	}

	actual := map[string]map[string]bool{}
	if err := json.Unmarshal(b, &actual); err != nil {
		t.Errorf("unexpected error unmarshaling ACL: %v\n", err)
		t.FailNow()
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("ACL did not marshal correct.\nGot:\n%v\nExpected:\n%v\n", actual, expected)
	}
}

func TestACLUnmarshal(t *testing.T) {
	b := `{"*":{"read":true},"abc":{"read":true},"def":{"read":true,"write":true},"role:xyz":{"read":true},"role:qrs":{"write":true,"read":true}}`

	acl := NewACL()
	if err := json.Unmarshal([]byte(b), &acl); err != nil {
		t.Errorf("unexpected error unmarshaling acl: %v\n", err)
		t.FailNow()
	}

	if !acl.PublicReadAccess() {
		t.Errorf("ACL does not have public read = true!")
	}

	if acl.PublicWriteAccess() {
		t.Errorf("ACL does has public write = true!")
	}

	cases := []struct {
		key           string
		isRole        bool
		expectedRead  bool
		expectedWrite bool
	}{
		{"abc", false, true, false},
		{"def", false, true, true},
		{"xyz", true, true, false},
		{"qrs", true, true, true},
		{"ghi", false, false, false},
		{"123", false, false, false},
		{"aaa", true, false, false},
		{"bbb", true, false, false},
	}

	for _, c := range cases {
		if c.isRole {
			if acl.RoleReadAccess(c.key) != c.expectedRead {
				t.Errorf("acl did not unmarshal correctly. Expected read=%v for role [%v], got %v\n", c.expectedRead, c.key, !c.expectedRead)
			}
			if acl.RoleWriteAccess(c.key) != c.expectedWrite {
				t.Errorf("acl did not unmarshal correctly. Expected write=%v for role [%v], got %v\n", c.expectedWrite, c.key, !c.expectedWrite)
			}
		} else {
			if acl.ReadAccess(c.key) != c.expectedRead {
				t.Errorf("acl did not unmarshal correctly. Expected read=%v for id [%v], got %v\n", c.expectedRead, c.key, !c.expectedRead)
			}
			if acl.WriteAccess(c.key) != c.expectedWrite {
				t.Errorf("acl did not unmarshal correctly. Expected write=%v for id [%v], got %v\n", c.expectedWrite, c.key, !c.expectedWrite)
			}
		}
	}
}

func TestConfig(t *testing.T) {
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"params":{"bool":true,"string":"blah blah blah","number":123.4,"object":{"a":false,"b":73}}}`)
	})
	defer teardownTestServer()

	c, err := GetConfig()
	if err != nil {
		t.Errorf("unexpected error on GetConfig: %v\n", err)
		t.FailNow()
	}

	expectedConf := Config{
		"bool":   true,
		"string": "blah blah blah",
		"number": 123.4,
		"object": map[string]interface{}{
			"a": false,
			"b": 73.0,
		},
	}

	if !reflect.DeepEqual(c, expectedConf) {
		t.Errorf("config was different from expected.\nGot:\n%v\nExpected:\n%v\n", c, expectedConf)
	}
}

type ClassNameTestType struct{}

type CustomClassNameTestType struct{}

func (c *CustomClassNameTestType) Endpoint() string {
	return "other/ep"
}
func (c *CustomClassNameTestType) ClassName() string {
	return "OtherName"
}

type CustomClassNameTestType2 struct{}

func (c CustomClassNameTestType2) Endpoint() string {
	return "other/ep2"
}
func (c CustomClassNameTestType2) ClassName() string {
	return "OtherName2"
}

func TestGetClassName(t *testing.T) {
	cases := []struct {
		inst     interface{}
		expected string
	}{
		{&ClassNameTestType{}, "ClassNameTestType"},
		{&CustomClassNameTestType{}, "OtherName"},
		{&CustomClassNameTestType2{}, "OtherName2"},
	}

	for _, tc := range cases {
		actual := getClassName(tc.inst)
		if actual != tc.expected {
			t.Errorf("Wrong class name returned for test case [%+v] - got [%s]\n", tc, actual)
		}
	}
}

func TestGetEndpointBase(t *testing.T) {
	cases := []struct {
		inst     interface{}
		expected string
	}{
		{&ClassNameTestType{}, "1/classes/ClassNameTestType"},
		{&[]ClassNameTestType{}, "1/classes/ClassNameTestType"},
		{&[]*ClassNameTestType{}, "1/classes/ClassNameTestType"},
		{&CustomClassNameTestType{}, "1/other/ep"},
		{&[]CustomClassNameTestType{}, "1/other/ep"},
		{&[]*CustomClassNameTestType{}, "1/other/ep"},
		{&CustomClassNameTestType2{}, "1/other/ep2"},
		{&[]CustomClassNameTestType2{}, "1/other/ep2"},
		{&[]*CustomClassNameTestType2{}, "1/other/ep2"},
	}

	for _, tc := range cases {
		actual := getEndpointBase(tc.inst)
		if actual != tc.expected {
			t.Errorf("Wrong endpoint name returned for test case [%+v] - got [%s]\n", tc, actual)
		}
	}
}

func TestRegisterType(t *testing.T) {
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"f1":{"__type":"Object","className":"CustType1","a":1,"b":"blah","c":true},"f2":{"__type":"Object","className":"CustType2","z":73.37,"y":"foobar","X":11}}`)
	})
	defer teardownTestServer()

	type CustType1 struct {
		A int
		B string
		C bool
	}

	type CustType2 struct {
		Z float64
		Y string
		X int
	}

	type TestType struct {
		F1 interface{}
		F2 interface{}
	}

	RegisterType(new(CustType1))
	RegisterType(new(CustType2))

	tt := TestType{}
	q, _ := NewQuery(&tt)
	if err := q.Get("123"); err != nil {
		t.Errorf("Unexpected error on Get: %v\n", err)
		t.FailNow()
	}

	if c1, ok := tt.F1.(*CustType1); ok {
		if c1.A != 1 {
			t.Errorf("CustType1.A value different from expected - expected 1, got [%v]\n", c1.A)
		}

		if c1.B != "blah" {
			t.Errorf("CustType1.B value different from expected - expected \"blah\", got [%q]\n", c1.B)
		}

		if c1.C != true {
			t.Errorf("CustType1.C value different from expected - expected true, got [%v]\n", c1.C)
		}
	} else {
		t.Errorf("Expected F1 to be of type *CustType1, got: %v\n", reflect.TypeOf(tt.F1))
	}

	if c2, ok := tt.F2.(*CustType2); ok {
		if c2.Z != 73.37 {
			t.Errorf("CustType2.Z value different from expected - expected 73.37, got [%v]\n", c2.Z)
		}

		if c2.Y != "foobar" {
			t.Errorf("CustType2.Y value different from expected - expected \"foobar\", got [%q]\n", c2.Y)
		}

		if c2.X != 11 {
			t.Errorf("CustType2.X value different from expected - expected 11, got [%v]\n", c2.X)
		}
	} else {
		t.Errorf("Expected F2 to be of type *CustType2, got: %v\n", reflect.TypeOf(tt.F1))
	}
}
