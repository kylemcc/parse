package parse

import "testing"

type CustomClass struct{
	Base
}

type CustomClassCustomName struct{
	Base
}

func (c CustomClassCustomName) ClassName() string {
	return "customName"
}

type CustomClassCustomEndpoint struct{
	Base
}

func (c *CustomClassCustomEndpoint) Endpoint() string {
	return "custom/class/endpoint"
}

func TestEndpoint(t *testing.T) {
	testCases := []struct {
		inst     interface{}
		expected string
	}{
		{&User{}, "https://api.parse.com/1/users"},
		{&CustomClass{}, "https://api.parse.com/1/classes/CustomClass"},
		{&CustomClassCustomName{}, "https://api.parse.com/1/classes/customName"},
		{&CustomClassCustomEndpoint{}, "https://api.parse.com/1/custom/class/endpoint"},
	}

	for _, tc := range testCases {
		q, err := NewQuery(tc.inst)
		if err != nil {
			t.Errorf("Unexpected error creating query: %v\n", err)
			t.FailNow()
		}
		qt := q.(*queryT)
		actual, err := qt.endpoint()
		if err != nil {
			t.Errorf("Unexpected error creating query: %v\n", err)
			continue
		}
		if actual != tc.expected {
			t.Errorf("Wrong endpoint generated. Expected [%s] got [%s]\n", tc.expected, actual)
		}
	}
}

func TestEndpointGetUpdateDelete(t *testing.T) {
	testCases := []struct {
		inst     interface{}
		id string
		expected string
	}{
		{&User{}, "UserId1", "https://api.parse.com/1/users/UserId1"},
		{&CustomClass{}, "Custom1", "https://api.parse.com/1/classes/CustomClass/Custom1"},
		{&CustomClassCustomName{}, "CC2", "https://api.parse.com/1/classes/customName/CC2"},
		{&CustomClassCustomEndpoint{}, "Cc3", "https://api.parse.com/1/custom/class/endpoint/Cc3"},
	}

	ops := []opTypeT{otGet, otDelete, otUpdate}

	for _, tc := range testCases {
		q, err := NewQuery(tc.inst)
		if err != nil {
			t.Errorf("Unexpected error creating query: %v\n", err)
			t.FailNow()
		}
		for _, ot := range ops {
			qt := q.(*queryT)
			qt.op = ot
			qt.instId = &tc.id
			actual, err := qt.endpoint()
			if err != nil {
				t.Errorf("Unexpected error creating query: %v\n", err)
				continue
			}
			if actual != tc.expected {
				t.Errorf("Wrong endpoint generated. Expected [%s] got [%s]\n", tc.expected, actual)
			}
		}
	}
}
