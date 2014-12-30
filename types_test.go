package parse

import (
	"encoding/json"
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
