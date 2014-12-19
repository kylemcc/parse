package parse

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

type CustomClass struct {
	Base
}

type CustomClassCustomName struct {
	Base
}

func (c CustomClassCustomName) ClassName() string {
	return "customName"
}

type CustomClassCustomEndpoint struct {
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
		id       string
		expected string
	}{
		{&User{}, "UserId1", "https://api.parse.com/1/users/UserId1"},
		{&CustomClass{}, "Custom1", "https://api.parse.com/1/classes/CustomClass/Custom1"},
		{&CustomClassCustomName{}, "CC2", "https://api.parse.com/1/classes/customName/CC2"},
		{&CustomClassCustomEndpoint{}, "Cc3", "https://api.parse.com/1/custom/class/endpoint/Cc3"},
	}

	ops := []opTypeT{otGet}

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

type TestType struct {
	F1 string
	F2 int
	F3 time.Time
	F4 User
	F5 *User
	F6 int
}

func TestFilters(t *testing.T) {
	q, err := NewQuery(&TestType{})
	if err != nil {
		t.Errorf("Uenexpected error creating query: %v\n", err)
		t.FailNow()
	}

	qt := q.(*queryT)

	q.EqualTo("f1", "test")
	q.EqualTo("f2", 1)
	q.EqualTo("f3", time.Date(2014, 1, 14, 13, 37, 6, 120000000, time.UTC))
	q.EqualTo("f4", "abcdefg")
	q.EqualTo("f5", User{Base{Id: "qrstuv"}})
	q.NotEqualTo("f6", 7)
	q.GreaterThan("f7", 3.2)
	q.GreaterThanOrEqual("f8", "abc")
	q.LessThan("f9", 73)
	q.LessThanOrEqual("f10", 789)
	q.In("f11", "abc", "def", "ghi")
	q.NotIn("f12", 7, 8, 9)
	q.Exists("f13")
	q.DoesNotExist("f14")
	q.All("f15", 1.1, 2.2, 3.3)
	q.Contains("f16", "substr")
	q.StartsWith("f17", "start")
	q.EndsWith("f18", "end")
	q.WithinGeoBox("f19", GeoPoint{41.9373658, -87.6746106}, GeoPoint{41.9414359, -87.6645255})
	q.Near("f20", GeoPoint{41.894303, -87.676835})
	q.WithinMiles("f21", GeoPoint{41.894303, -87.676835}, 5)
	q.WithinKilometers("f22", GeoPoint{41.894303, -87.676835}, 7.8)
	q.WithinRadians("f23", GeoPoint{41.894303, -87.676835}, 0.8910)

	em := map[string]interface{}{
		"f1": "test",
		"f2": 1,
		"f3": map[string]interface{}{
			"__type": "Date",
			"iso":    "2014-01-14T13:37:06.120Z",
		},
		"f4": map[string]interface{}{
			"__type":    "Pointer",
			"className": "_User",
			"objectId":  "abcdefg",
		},
		"f5": map[string]interface{}{
			"__type":    "Pointer",
			"className": "_User",
			"objectId":  "qrstuv",
		},
		"f6": map[string]interface{}{
			"$ne": 7,
		},
		"f7": map[string]interface{}{
			"$gt": 3.2,
		},
		"f8": map[string]interface{}{
			"$gte": "abc",
		},
		"f9": map[string]interface{}{
			"$lt": 73,
		},
		"f10": map[string]interface{}{
			"$lte": 789,
		},
		"f11": map[string]interface{}{
			"$in": []interface{}{"abc", "def", "ghi"},
		},
		"f12": map[string]interface{}{
			"$nin": []interface{}{7, 8, 9},
		},
		"f13": map[string]interface{}{
			"$exists": true,
		},
		"f14": map[string]interface{}{
			"$exists": false,
		},
		"f15": map[string]interface{}{
			"$all": []interface{}{1.1, 2.2, 3.3},
		},
		"f16": map[string]interface{}{
			"$regex": "\\Qsubstr\\E",
		},
		"f17": map[string]interface{}{
			"$regex": "^\\Qstart\\E",
		},
		"f18": map[string]interface{}{
			"$regex": "\\Qend\\E$",
		},
		"f19": map[string]interface{}{
			"$within": map[string]interface{}{
				"$box": []map[string]interface{}{
					map[string]interface{}{
						"__type":    "GeoPoint",
						"latitude":  41.9373658,
						"longitude": -87.6746106,
					},
					map[string]interface{}{
						"__type":    "GeoPoint",
						"latitude":  41.9414359,
						"longitude": -87.6645255,
					},
				},
			},
		},
		"f20": map[string]interface{}{
			"$nearSphere": map[string]interface{}{
				"__type":    "GeoPoint",
				"latitude":  41.894303,
				"longitude": -87.676835,
			},
		},
		"f21": map[string]interface{}{
			"$nearSphere": map[string]interface{}{
				"__type":    "GeoPoint",
				"latitude":  41.894303,
				"longitude": -87.676835,
			},
			"$maxDistanceInMiles": 5,
		},
		"f22": map[string]interface{}{
			"$nearSphere": map[string]interface{}{
				"__type":    "GeoPoint",
				"latitude":  41.894303,
				"longitude": -87.676835,
			},
			"$maxDistanceInKilometers": 7.8,
		},
		"f23": map[string]interface{}{
			"$nearSphere": map[string]interface{}{
				"__type":    "GeoPoint",
				"latitude":  41.894303,
				"longitude": -87.676835,
			},
			"$maxDistanceInRadians": 0.8910,
		},
	}

	expected := map[string]interface{}{}
	eb, _ := json.Marshal(&em)
	_ = json.Unmarshal(eb, &expected)

	b, err := json.Marshal(&qt.where)
	if err != nil {
		t.Errorf("error marshaling where: %v\n", err)
		t.FailNow()
	}

	actual := map[string]interface{}{}
	err = json.Unmarshal(b, &actual)
	if err != nil {
		t.Errorf("error unmarshaling where: %v\n", err)
		t.FailNow()
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("where different from expected. expected:\n%s\n\ngot:\n%s\n", eb, b)
	}
}
