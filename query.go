package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type opTypeT int

const (
	otInval opTypeT = iota
	otGet
	otQuery
)

type Query interface {

	// Use the Master Key for the given request.
	UseMasterKey() Query

	// Get retrieves the instance of the type pointed to by v and
	// identified by id, and stores the result in v.
	Get(id string) error

	// Set the sort order for the query. The first argument sets the primary
	// sort order. Subsequent arguments will set secondary sort orders. Results
	// will be sorted in ascending order by default. Prefix field names with a
	// '-' to sort in descending order. E.g.: q.OrderBy("-createdAt") will sort
	// by the createdAt field in descending order.
	OrderBy(fs ...string) Query

	// Set the number of results to retrieve
	Limit(l int) Query

	// Set the number of results to skip before returning any results
	Skip(s int) Query

	// Specify nested fields to retrieve within the primary object. Use
	// dot notation to retrieve further nested fields. E.g.:
	// q.Include("user") or q.Include("user.location")
	Include(fs ...string) Query

	// Only retrieve the specified fields
	Keys(fs ...string) Query

	// Add a constraint requiring the field specified by f be equal to the
	// value represented by v
	EqualTo(f string, v interface{}) Query

	// Add a constraint requiring the field specified by f not be equal to the
	// value represented by v
	NotEqualTo(f string, v interface{}) Query

	// Add a constraint requiring the field specified by f be greater than the
	// value represented by v
	GreaterThan(f string, v interface{}) Query

	// Add a constraint requiring the field specified by f be greater than or
	// or equal to the value represented by v
	GreaterThanOrEqual(f string, v interface{}) Query

	// Add a constraint requiring the field specified by f be less than the
	// value represented by v
	LessThan(f string, v interface{}) Query

	// Add a constraint requiring the field specified by f be less than or
	// or equal to the value represented by v
	LessThanOrEqual(f string, v interface{}) Query

	// Add a constraint requiring the field specified by f be equal to one
	// of the values specified
	In(f string, vs ...interface{}) Query

	// Add a constraint requiring the field specified by f not be equal to any
	// of the values specified
	NotIn(f string, vs ...interface{}) Query

	// Add a constraint requiring returned objects contain the field specified by f
	Exists(f string) Query

	// Add a constraint requiring returned objects do not contain the field specified by f
	DoesNotExist(f string) Query

	// Add a constraint requiring the field specified by f contain all
	// of the values specified
	All(f string, vs ...interface{}) Query

	// Add a constraint requiring the string field specified by f contain
	// the substring specified by v
	Contains(f string, v string) Query

	// Add a constraint requiring the string field specified by f start with
	// the substring specified by v
	StartsWith(f string, v string) Query

	// Add a constraint requiring the string field specified by f end with
	// the substring specified by v
	EndsWith(f string, v string) Query

	// Add a constraint requiring the string field specified by f match the
	// regular expression v
	Matches(f string, v string, ignoreCase bool, multiLine bool) Query

	// Add a constraint requiring the location of GeoPoint field specified by f be
	// within the rectangular geographic bounding box with a southwest corner
	// represented by sw and a northeast corner represented by ne
	WithinGeoBox(f string, sw GeoPoint, ne GeoPoint) Query

	// Add a constraint requiring the location of GeoPoint field specified by f
	// be near the point represented by g
	Near(f string, g GeoPoint) Query

	// Add a constraint requiring the location of GeoPoint field specified by f
	// be near the point represented by g with a maximum distance in miles
	// represented by m
	WithinMiles(f string, g GeoPoint, m float64) Query

	// Add a constraint requiring the location of GeoPoint field specified by f
	// be near the point represented by g with a maximum distance in kilometers
	// represented by m
	WithinKilometers(f string, g GeoPoint, k float64) Query

	// Add a constraint requiring the location of GeoPoint field specified by f
	// be near the point represented by g with a maximum distance in radians
	// represented by m
	WithinRadians(f string, g GeoPoint, r float64) Query

	// Add a constraint requiring the value of the field specified by f be equal
	// to the field named qk in the result of the subquery sq
	MatchesKeyInQuery(f string, qk string, sq Query) Query

	// Add a constraint requiring the value of the field specified by f not match
	// the field named qk in the result of the subquery sq
	DoesNotMatchKeyInQuery(f string, qk string, sq Query) Query

	// Constructs a query where each result must satisfy one of the given
	// subueries
	//
	// E.g.:
	//
	// q, _ := parse.NewQuery(&parse.User{})
	//
	// sq1, _ := parse.NewQuery(&parse.User{})
	// sq1.EqualTo("city", "Chicago")
	//
	// sq2, _ := parse.NewQuery(&parse.User{})
	// sq2.GreaterThan("age", 30)
	//
	// sq3, _ := parse.NewQuery(&parse.User{})
	// sq3.In("occupation", []string{"engineer", "developer"})
	//
	// q.Or(sq1, sq2, sq3)
	// q.Each(...)
	Or(qs ...Query) Query

	// Iterate of each result of a query, passing the result to the provided
	// channel rc. Errors are passed to the channel ec
	Each(rc interface{}, ec chan<- error) error

	// Retrieves a list of objects that satisfy the given query. The results
	// are assigned to the slice provided to NewQuery.
	//
	// E.g.:
	//
	// users := make([]parse.User)
	// q, _ := parse.NewQuery(&users)
	// q.EqualTo("city", "Chicago")
	// q.OrderBy("-createdAt")
	// q.Limit(20)
	// q.Find() // Retrieve the 20 newest users in Chicago
	Find() error

	// Retrieves the first result that satisfies the given query. The result
	// is assigned to the value provided to NewQuery.
	//
	// E.g.:
	// u := parse.User{}
	// q, _ := parse.NewQuery(&u)
	// q.EqualTo("city", "Chicago")
	// q.OrderBy("-createdAt")
	// q.First() // Retrieve the newest user in Chicago
	First() error

	// Retrieve the number of results that satisfy the given query
	Count() (int64, error)
}

type queryT struct {
	inst interface{}
	op   opTypeT

	instId    *string
	orderBy   []string
	limit     *int
	skip      *int
	count     *int
	where     map[string]interface{}
	include   map[string]struct{}
	keys      map[string]struct{}
	className string

	currentSession *sessionT

	shouldUseMasterKey bool
}

// Create a new query instance.
func NewQuery(v interface{}) (Query, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, errors.New("v must be a non-nil pointer")
	}

	return &queryT{
		inst:      v,
		orderBy:   make([]string, 0),
		where:     make(map[string]interface{}),
		include:   make(map[string]struct{}),
		keys:      make(map[string]struct{}),
		className: getClassName(v),
	}, nil
}

func (q *queryT) UseMasterKey() Query {
	q.shouldUseMasterKey = true
	return q
}

func (q *queryT) Get(id string) error {
	q.op = otGet
	q.instId = &id
	return defaultClient.doRequest(q, q.inst)
}

func (q *queryT) OrderBy(fs ...string) Query {
	q.orderBy = append(make([]string, 0, len(fs)), fs...)
	return q
}

func (q *queryT) Limit(l int) Query {
	q.limit = &l
	return q
}

func (q *queryT) Skip(s int) Query {
	q.skip = &s
	return q
}

func (q *queryT) Include(fs ...string) Query {
	for _, f := range fs {
		q.include[f] = struct{}{}
	}
	return q
}

func (q *queryT) Keys(fs ...string) Query {
	for _, f := range fs {
		q.keys[f] = struct{}{}
	}
	return q
}

func (q *queryT) EqualTo(f string, v interface{}) Query {
	qv := getQueryRepr(q.inst, f, v)
	q.where[f] = qv
	return q
}

func (q *queryT) NotEqualTo(f string, v interface{}) Query {
	qv := getQueryRepr(q.inst, f, v)
	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$ne"] = qv
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$ne": qv,
	}
	return q
}

func (q *queryT) GreaterThan(f string, v interface{}) Query {
	var qv interface{}
	if t, ok := v.(time.Time); ok {
		qv = Date(t)
	} else if t, ok := v.(*time.Time); ok {
		qv = Date(*t)
	} else {
		qv = v
	}

	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$gt"] = qv
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$gt": qv,
	}
	return q
}

func (q *queryT) GreaterThanOrEqual(f string, v interface{}) Query {
	var qv interface{}
	if t, ok := v.(time.Time); ok {
		qv = Date(t)
	} else if t, ok := v.(*time.Time); ok {
		qv = Date(*t)
	} else {
		qv = v
	}

	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$gte"] = qv
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$gte": qv,
	}
	return q
}

func (q *queryT) LessThan(f string, v interface{}) Query {
	var qv interface{}
	if t, ok := v.(time.Time); ok {
		qv = Date(t)
	} else if t, ok := v.(*time.Time); ok {
		qv = Date(*t)
	} else {
		qv = v
	}

	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$lt"] = qv
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$lt": qv,
	}
	return q
}

func (q *queryT) LessThanOrEqual(f string, v interface{}) Query {
	var qv interface{}
	if t, ok := v.(time.Time); ok {
		qv = Date(t)
	} else if t, ok := v.(*time.Time); ok {
		qv = Date(*t)
	} else {
		qv = v
	}

	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$lte"] = qv
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$lte": qv,
	}
	return q
}

func (q *queryT) In(f string, vs ...interface{}) Query {
	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$in"] = vs
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$in": vs,
	}
	return q
}

func (q *queryT) NotIn(f string, vs ...interface{}) Query {
	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$nin"] = vs
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$nin": vs,
	}
	return q
}

func (q *queryT) Exists(f string) Query {
	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$exists"] = true
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$exists": true,
	}
	return q
}

func (q *queryT) DoesNotExist(f string) Query {
	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$exists"] = false
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$exists": false,
	}
	return q
}

func (q *queryT) All(f string, vs ...interface{}) Query {
	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$all"] = vs
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$all": vs,
	}
	return q
}

func (q *queryT) Contains(f string, v string) Query {
	v = quote(v)
	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$regex"] = v
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$regex": v,
	}
	return q
}

func (q *queryT) StartsWith(f string, v string) Query {
	v = "^" + quote(v)
	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$regex"] = v
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$regex": v,
	}
	return q
}

func (q *queryT) EndsWith(f string, v string) Query {
	v = quote(v) + "$"
	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$regex"] = v
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$regex": v,
	}
	return q
}

func (q *queryT) Matches(f string, v string, ignoreCase bool, multiLine bool) Query {
	v = quote(v)
	if cv, ok := q.where[f]; ok {
		if m, ok := cv.(map[string]interface{}); ok {
			m["$regex"] = v
			return q
		}
	}

	q.where[f] = map[string]interface{}{
		"$regex": v,
	}

	var options string

	if ignoreCase {
		options += "i"
	}

	if multiLine {
		options += "m"
	}

	if len(options) > 0 {
		if m, ok := q.where[f].(map[string]interface{}); ok {
			m["$options"] = options
		}
	}

	return q
}

func (q *queryT) WithinGeoBox(f string, sw GeoPoint, ne GeoPoint) Query {
	q.where[f] = map[string]interface{}{
		"$within": map[string]interface{}{
			"$box": []GeoPoint{sw, ne},
		},
	}
	return q
}

func (q *queryT) Near(f string, g GeoPoint) Query {
	q.where[f] = map[string]interface{}{
		"$nearSphere": g,
	}
	return q
}

func (q *queryT) WithinMiles(f string, g GeoPoint, m float64) Query {
	q.where[f] = map[string]interface{}{
		"$nearSphere":         g,
		"$maxDistanceInMiles": m,
	}
	return q
}

func (q *queryT) WithinKilometers(f string, g GeoPoint, k float64) Query {
	q.where[f] = map[string]interface{}{
		"$nearSphere":              g,
		"$maxDistanceInKilometers": k,
	}
	return q
}

func (q *queryT) WithinRadians(f string, g GeoPoint, r float64) Query {
	q.where[f] = map[string]interface{}{
		"$nearSphere":           g,
		"$maxDistanceInRadians": r,
	}
	return q
}

func (q *queryT) MatchesKeyInQuery(f, qk string, sq Query) Query {
	var sqt *queryT
	if tmp, ok := sq.(*queryT); ok {
		sqt = tmp
	}

	q.where[f] = map[string]interface{}{
		"$select": map[string]interface{}{
			"key":   qk,
			"query": sqt,
		},
	}
	return q
}

func (q *queryT) DoesNotMatchKeyInQuery(f string, qk string, sq Query) Query {
	var sqt *queryT
	if tmp, ok := sq.(*queryT); ok {
		sqt = tmp
	}

	q.where[f] = map[string]interface{}{
		"$dontSelect": map[string]interface{}{
			"key":   qk,
			"query": sqt,
		},
	}
	return q
}

func (q *queryT) Or(qs ...Query) Query {
	or := make([]map[string]interface{}, 0, len(qs))
	for _, qi := range qs {
		if qt, ok := qi.(*queryT); ok {
			or = append(or, qt.where)
		}
	}
	q.where["$or"] = or
	return q
}

func (q *queryT) Each(rc interface{}, ec chan<- error) error {
	rv := reflect.ValueOf(rc)
	rt := rv.Type()
	if rt.Kind() != reflect.Chan {
		return fmt.Errorf("rc must be a channel, received %s", rt.Kind())
	}

	if rt.Elem().Kind() == reflect.Ptr {
		if rt.Elem() != reflect.TypeOf(q.inst) {
			return fmt.Errorf("1rc must be of type chan %s, received chan %s", reflect.TypeOf(q.inst), rt.Elem())
		}
	} else {
		if rt.Elem() != reflect.TypeOf(q.inst).Elem() {
			return fmt.Errorf("2rc must be of type chan %s, received chan %s", reflect.TypeOf(q.inst).Elem(), rt.Elem())
		}
	}

	if q.op == otInval {
		q.op = otQuery
	}

	if q.limit != nil || q.skip != nil || len(q.orderBy) > 0 {
		return errors.New("cannot iterate over a query with a sort, limit, or skip")
	}

	q.OrderBy("objectId")
	q.Limit(100)

	go func() {
		for {
			s := reflect.New(reflect.SliceOf(rt.Elem()))
			s.Elem().Set(reflect.MakeSlice(reflect.SliceOf(rt.Elem()), 0, 100))

			err := defaultClient.doRequest(q, s.Interface())
			if err != nil {
				ec <- err
			}

			for i := 0; i < s.Elem().Len(); i++ {
				rv.Send(s.Elem().Index(i))
			}

			if s.Elem().Len() < *q.limit {
				break
			} else {
				last := s.Elem().Index(s.Elem().Len() - 1)
				last = reflect.Indirect(last)
				if f := last.FieldByName("Id"); f.IsValid() {
					if id, ok := f.Interface().(string); ok {
						q.GreaterThan("objectId", id)
					}
				}

			}
		}
		rv.Close()
		close(ec)
	}()

	return nil
}

func (q *queryT) Find() error {
	q.op = otQuery
	return defaultClient.doRequest(q, q.inst)
}

func (q *queryT) First() error {
	q.op = otQuery
	l := 1
	q.limit = &l
	return defaultClient.doRequest(q, q.inst)
}

func (q *queryT) Count() (int64, error) {
	l := 0
	c := 1
	q.limit = &l
	q.count = &c

	var count int64
	err := defaultClient.doRequest(q, &count)
	return count, err
}

func (q *queryT) payload() (string, error) {
	p := url.Values{}
	if len(q.where) > 0 {
		w, err := json.Marshal(q.where)
		if err != nil {
			return "", err
		}
		p["where"] = []string{string(w)}
	}

	if q.limit != nil {
		p["limit"] = []string{strconv.Itoa(*q.limit)}
	}

	if q.skip != nil {
		p["skip"] = []string{strconv.Itoa(*q.skip)}
	}

	if q.count != nil {
		p["count"] = []string{strconv.Itoa(*q.count)}
	}

	if len(q.orderBy) > 0 {
		o := strings.Join(q.orderBy, ",")
		p["order"] = []string{o}
	}

	if len(q.include) > 0 {
		is := make([]string, 0, len(q.include))
		for k := range q.include {
			is = append(is, k)
		}
		i := strings.Join(is, ",")
		p["include"] = []string{i}
	}

	if len(q.keys) > 0 {
		ks := make([]string, 0, len(q.include))
		for k := range q.keys {
			ks = append(ks, k)
		}
		k := strings.Join(ks, ",")
		p["keys"] = []string{k}
	}

	return p.Encode(), nil
}

// Implement the operationT interface
func (q *queryT) method() string {
	return "GET"
}

func (q *queryT) endpoint() (string, error) {
	u := url.URL{}
	p := getEndpointBase(q.inst)

	switch q.op {
	case otGet:
		p = path.Join(p, *q.instId)
	}

	qs, err := q.payload()
	if err != nil {
		return "", err
	}

	u.Scheme = "https"
	u.Host = parseHost
	u.RawQuery = qs
	u.Path = p

	return u.String(), nil
}

func (q *queryT) body() (string, error) {
	return "", nil
}

func (q *queryT) useMasterKey() bool {
	return q.shouldUseMasterKey
}

func (q *queryT) session() *sessionT {
	return q.currentSession
}

func (q *queryT) contentType() string {
	return "application/x-www-form-urlencoded"
}

func (q *queryT) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}

	if len(q.where) > 0 {
		m["where"] = q.where
	}

	if q.className != "" {
		m["className"] = q.className
	}

	if q.limit != nil {
		m["limit"] = q.limit
	}

	if q.skip != nil {
		m["skip"] = q.skip
	}

	if len(q.orderBy) > 0 {
		m["skip"] = q.orderBy
	}

	if len(q.include) > 0 {
		m["include"] = q.include
	}

	if len(q.keys) > 0 {
		m["keys"] = q.keys
	}

	return json.Marshal(m)
}

// From the Javascript library - convert the string represented by re into a regex
// value that matches it. MongoDb (what backs Parse) uses PCRE syntax
func quote(re string) string {
	return "\\Q" + strings.Replace(re, "\\E", "\\E\\\\E\\Q", -1) + "\\E"
}

func getQueryRepr(inst interface{}, f string, v interface{}) interface{} {
	var fname string
	fieldMap := getFieldNameMap(reflect.ValueOf(inst))

	if fn, ok := fieldMap[f]; ok {
		fname = fn
	} else {
		fname = f
	}
	fname = firstToUpper(fname)

	rvInst := reflect.ValueOf(inst)
	rviInst := reflect.Indirect(rvInst)
	rtInst := rviInst.Type()
	if rtInst.Kind() == reflect.Slice || rtInst.Kind() == reflect.Array {
		rtInst = rtInst.Elem()
		if rtInst.Kind() == reflect.Ptr {
			rtInst = rtInst.Elem()
		}
	}

	if rtInst.Kind() == reflect.Struct {
		if sf, ok := rtInst.FieldByName(fname); ok {
			ft := sf.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}

			if ft.Kind() == reflect.Struct {
				if ft == reflect.TypeOf(time.Time{}) || ft == reflect.TypeOf(Date{}) {
					switch v.(type) {
					case time.Time:
						return Date(v.(time.Time))
					case *time.Time:
						return Date(*v.(*time.Time))
					case Date, *Date:
						return v
					case string:
						return map[string]interface{}{
							"__type": "Date",
							"iso":    v.(string),
						}
					}
				} else {
					var id string
					var cname string
					ftInst := reflect.Zero(ft)

					if tmp, ok := ftInst.Interface().(iClassName); ok {
						cname = tmp.ClassName()
					} else if tmp, ok := reflect.New(ft).Interface().(iClassName); ok {
						cname = tmp.ClassName()
					} else {
						cname = ft.Name()
					}

					rv := reflect.ValueOf(v)
					rvi := reflect.Indirect(rv)
					if rvi.Kind() == reflect.String {
						id = rvi.Interface().(string)
					} else if idf := rvi.FieldByName("Id"); idf.IsValid() {
						id = idf.Interface().(string)
					}

					return Pointer{
						Id:        id,
						ClassName: cname,
					}
				}
			}
		}
	}

	return v
}
