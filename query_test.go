package parse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

func TestEndpointGet(t *testing.T) {
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
	q.EqualTo("f5", User{Base: Base{Id: "qrstuv"}})
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

	subq, _ := NewQuery(&User{})
	subq.EqualTo("email", "kylemcc@gmail.com")
	q.MatchesKeyInQuery("f24", "testKey", subq)

	subq2, _ := NewQuery(&User{})
	subq2.EqualTo("city", "Chicago")
	q.DoesNotMatchKeyInQuery("f25", "testKey2", subq2)

	q.MatchesQuery("f26", subq)
	q.DoesNotMatchQuery("f27", subq2)

	q.Limit(10)
	q.Skip(20)
	q.OrderBy("-createdAt")
	q.Include("location")
	q.Keys("email")

	em := map[string]interface{}{
		"f1": "test",
		"f2": 1,
		"f3": map[string]interface{}{
			"__type": "Date",
			"iso":    "2014-01-14T13:37:06.120Z",
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
					{
						"__type":    "GeoPoint",
						"latitude":  41.9373658,
						"longitude": -87.6746106,
					},
					{
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
		"f24": map[string]interface{}{
			"$select": map[string]interface{}{
				"key": "testKey",
				"query": map[string]interface{}{
					"className": "_User",
					"where": map[string]interface{}{
						"email": "kylemcc@gmail.com",
					},
				},
			},
		},
		"f25": map[string]interface{}{
			"$dontSelect": map[string]interface{}{
				"key": "testKey2",
				"query": map[string]interface{}{
					"className": "_User",
					"where": map[string]interface{}{
						"city": "Chicago",
					},
				},
			},
		},
		"f26": map[string]interface{}{
			"$inQuery": map[string]interface{}{
				"className": "_User",
				"where": map[string]interface{}{
					"email": "kylemcc@gmail.com",
				},
			},
		},
		"f27": map[string]interface{}{
			"$notInQuery": map[string]interface{}{
				"className": "_User",
				"where": map[string]interface{}{
					"city": "Chicago",
				},
			},
		},
	}

	expected := map[string]interface{}{}
	eb, _ := json.Marshal(em)
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

	p, _ := q.(*queryT).payload()
	qs, err := url.ParseQuery(p)
	if err != nil {
		t.Errorf("unexpected error parsing query string: %v\n", err)
		t.FailNow()
	}

	cases := []struct {
		key      string
		expected string
	}{
		{"limit", "10"},
		{"skip", "20"},
		{"order", "-createdAt"},
		{"include", "location"},
		{"keys", "email"},
	}

	for _, c := range cases {
		if v := qs.Get(c.key); v != c.expected {
			t.Errorf("query value for key [%s] did not match. Got [%v] expected [%v]\n", c.key, v, c.expected)
		}
	}
}

func TestQueryRequiresPointer(t *testing.T) {
	u := User{}
	expected := "v must be a non-nil pointer"
	if _, err := NewQuery(u); err == nil {
		t.Error("NewQuery should return an error when argument is not a pointer")
	} else if err.Error() != expected {
		t.Errorf("Unexpected error message. Got [%s] expected [%s]\n", err, expected)
	}
}

func TestQueryRequest(t *testing.T) {
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

		r.ParseForm()
		whereStr := r.Form.Get("where")
		where := map[string]interface{}{}
		err := json.Unmarshal([]byte(whereStr), &where)
		if err != nil {
			t.Errorf("unexpected error unmarshaling where: %v\n", err)
			t.FailNow()
		}

		ew := map[string]interface{}{
			"city": "Chicago",
			"age": map[string]interface{}{
				"$gt": 30,
			},
		}
		ewb, err := json.Marshal(ew)
		if err != nil {
			t.Errorf("unexpected error unmarshaling expected where: %v\n", err)
			t.FailNow()
		}
		expected := map[string]interface{}{}
		err = json.Unmarshal(ewb, &expected)

		if !reflect.DeepEqual(expected, where) {
			t.Errorf("query where argument different from expected. Got [%s] expected [%s]\n", whereStr, ewb)
			t.FailNow()
		}

		fmt.Fprintf(w, `{"results":[{"objectId": "123", "createdAt":"2012-04-14T19:23:10.123Z"}]}`)
	})
	defer teardownTestServer()

	u := User{}
	q, err := NewQuery(&u)
	if err != nil {
		t.Errorf("Unexpected error creating query: %v\n", err)
		t.FailNow()
	}

	q.EqualTo("city", "Chicago")
	q.GreaterThan("age", 30)
	err = q.First()
	if err != nil {
		t.Errorf("Error running query: %v\n", err)
	}

	if u.Id != "123" {
		t.Errorf("Query did not fill in u.Id")
	}
}

func TestQueryUseMasterKey(t *testing.T) {
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
		fmt.Fprintf(w, `{"results":[{"objectId": "123", "createdAt":"2012-04-14T19:23:10.123Z"}]}`)
	})
	defer teardownTestServer()

	u := User{}
	q, err := NewQuery(&u)
	if err != nil {
		t.Errorf("Unexpected error creating query: %v\n", err)
		t.FailNow()
	}

	q.EqualTo("city", "Chicago").GreaterThan("age", 30).UseMasterKey()
	err = q.First()
	if err != nil {
		t.Errorf("Error running query: %v\n", err)
	}
}

func TestFirst(t *testing.T) {
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.Form.Get("limit") != "1" {
			t.Errorf("limit was not 1. got [%v]\n", r.Form.Get("limit"))
		}
		fmt.Fprintf(w, `{"results":[{"objectId": "123", "createdAt":"2012-04-14T19:23:10.123Z"}]}`)
	})
	defer teardownTestServer()

	u := User{}
	q, err := NewQuery(&u)
	if err != nil {
		t.Errorf("Unexpected error creating query: %v\n", err)
		t.FailNow()
	}

	q.EqualTo("city", "Chicago").GreaterThan("age", 30).UseMasterKey()
	err = q.First()
	if err != nil {
		t.Errorf("Error running query: %v\n", err)
	}

	if u.Id != "123" {
		t.Errorf("Query did not populate struct with correct value. Got: %v\n", u.Id)
	}

	us := make([]User, 0, 1)
	q2, err := NewQuery(&us)
	if err != nil {
		t.Errorf("Unexpected error creating query: %v\n", err)
		t.FailNow()
	}
	err = q2.First()
	if err != nil {
		t.Errorf("Error running query: %v\n", err)
	}

	if len(us) != 1 {
		t.Errorf("Query did not populate slice with correct number of values. Len: %v\n", len(us))
		t.FailNow()
	}

	if us[0].Id != "123" {
		t.Errorf("Query did not populate struct with correct value. Got: %v\n", us[0].Id)
	}
}

func TestCount(t *testing.T) {
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		fmt.Fprintf(w, `{"results":[],"count":73}`)
	})
	defer teardownTestServer()

	q, err := NewQuery(&User{})
	if err != nil {
		t.Errorf("Unexpected error creating query: %v\n", err)
		t.FailNow()
	}

	q.EqualTo("city", "Chicago")
	cnt, err := q.Count()
	if err != nil {
		t.Errorf("Error running query: %v\n", err)
	}

	if cnt != 73 {
		t.Errorf("Count returned incorrect value. Got [%d] expected [%d]\n", cnt, 73)
	}
}

func TestFind(t *testing.T) {
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"results":[{"objectId": "123", "createdAt":"2012-04-14T19:23:10.123Z"},{"objectId":"abc","createdAt":"2012-04-14T19:23:10.123Z"}]}`)
	})
	defer teardownTestServer()

	us := make([]User, 0, 1)
	q, err := NewQuery(&us)
	if err != nil {
		t.Errorf("Unexpected error creating query: %v\n", err)
		t.FailNow()
	}
	err = q.Find()
	if err != nil {
		t.Errorf("Error running query: %v\n", err)
	}

	expectedIds := []string{"123", "abc"}
	if len(us) != len(expectedIds) {
		t.Errorf("Find returned the wrong number of results. Got [%d] expected [%d]\n", len(us), len(expectedIds))
	}

	for i, v := range expectedIds {
		if v != us[i].Id {
			t.Errorf("Find did not return proper result at index %d: Got [%v] expected [%v]\n", i, us[i].Id, v)
		}
	}
}

func TestGet(t *testing.T) {
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/users/abc123" {
			t.Errorf("Get requested wrong path. Got [%s] expected [%s]\n", r.URL.Path, "/1/users/abc123")
		}
		fmt.Fprintf(w, `{"objectId":"abc123","createdAt":"2012-04-14T19:23:10.123Z"}`)
	})
	defer teardownTestServer()

	u := User{}
	q, err := NewQuery(&u)
	if err != nil {
		t.Errorf("Unexpected error creating query: %v\n", err)
		t.FailNow()
	}
	err = q.Get("abc123")
	if err != nil {
		t.Errorf("Error running query: %v\n", err)
	}

	if u.Id != "abc123" {
		t.Errorf("Get returned wrong Id. Got: %v\n", u.Id)
	}
}

func TestEach(t *testing.T) {
	numRequests := 0
	setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		r.ParseForm()
		if r.Form.Get("limit") != "100" {
			t.Errorf("Did not get proper limit. Expected 100 got [%s]\n", r.Form.Get("limit"))
		}
		if r.Form.Get("order") != "objectId" {
			t.Errorf("Did not get proper order. Expected objectId got [%s]\n", r.Form.Get("order"))
		}

		ret := make([]map[string]interface{}, 0, 100)
		if where := r.Form.Get("where"); where == "" {
			for i := 0; i < 100; i++ {
				ret = append(ret, map[string]interface{}{"objectId": string(rune(i + 65)), "createdAt": "2014-12-19T22:22:22.123Z"})
			}
		} else {
			for i := 0; i < 50; i++ {
				ret = append(ret, map[string]interface{}{"objectId": string(rune(i + 200)), "createdAt": "2014-12-19T22:22:22.123Z"})
			}
		}
		j, _ := json.Marshal(map[string]interface{}{"results": ret})
		fmt.Fprintf(w, string(j))
	})
	defer teardownTestServer()

	q, err := NewQuery(&User{})
	if err != nil {
		t.Errorf("Unexpected error creating query: %v\n", err)
		t.FailNow()
	}

	rc := make(chan *User)

	it, err := q.Each(rc)
	if err != nil {
		t.Errorf("Unexpected error executing each: %v\n", err)
		t.FailNow()
	}

	users := make([]*User, 0)
	errors := make([]error, 0)
loop:
	for {
		select {
		case u := <-rc:
			if u != nil {
				users = append(users, u)
			}
		case err := <-it.Done():
			if err != nil {
				errors = append(errors, err)
			}
			break loop
		}
	}

	if numRequests != 2 {
		t.Errorf("Each did not execute the expected number of requests. Expected 2, got: %d\n", numRequests)
	}

	if len(errors) != 0 {
		t.Errorf("Errors received from Query.Each: %v\n", errors)
	}

	if len(users) != 150 {
		t.Errorf("Wrong number of users received. Expected 150, got: %d\n", len(users))
	}
}

func TestGetQueryRepr(t *testing.T) {
	_ = &struct {
		F1 string
		F2 int
		F3 float32
		F4 bool
		F5 time.Time
		F6 *User
	}{}

	cases := []struct {
		v        interface{}
		fname    string
		expected interface{}
	}{
		{"string", "f1", "string"},
		{73, "f2", 73}, //int
		{14.3, "f3", 14.3},
		{true, "f4", true},
		{
			time.Date(2014, 12, 19, 16, 47, 23, 120000000, time.UTC),
			"f5",
			Date(time.Date(2014, 12, 19, 16, 47, 23, 120000000, time.UTC)),
		},
		{
			&User{Base: Base{Id: "abc"}},
			"f7",
			Pointer{
				ClassName: "_User",
				Id:        "abc",
			},
		},
	}

	for _, c := range cases {
		actual := encodeForRequest(c.v)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("getQueryRepr did not return expected value for field [%v]. Got: [%v] expected [%v]\n", c.fname, actual, c.expected)
		}
	}
}
