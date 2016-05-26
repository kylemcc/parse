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
	"sync"
	"time"
)

type opTypeT int

const (
	otInval opTypeT = iota
	otGet
	otQuery
)

// Returned when a query returns no results
var ErrNoRows = errors.New("no results returned")

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

	// Add a constraint requiring the field specified by f contain the object
	// returned by Parse query q
	MatchesQuery(f string, q Query) Query

	// Add a constraint requiring the field specified by f not contain the object
	// returned by the Parse query q
	DoesNotMatchQuery(f string, q Query) Query

	// Convenience method for duplicating a query
	Clone() Query

	// Convenience method for building a subquery for use with Query.Or
	Sub() Query

	// Constructs a query where each result must satisfy one of the given
	// subueries
	//
	// E.g.:
	//
	// q, _ := parse.NewQuery(&parse.User{})
	//
	// sq1 := q.Sub().EqualTo("city", "Chicago")
	//
	// sq2 := q.Sub().GreaterThan("age", 30)
	//
	// sq3 := q.Sub().In("occupation", []string{"engineer", "developer"})
	//
	// q.Or(sq1, sq2, sq3)
	// q.Each(...)
	Or(qs ...Query) Query

	// Fetch all results for a query, sending each result to the provided
	// channel rc. The element type of rc should match that of the query,
	// otherwise an error will be returned.
	//
	// Errors are passed to the channel ec. If an error occurns during iteration,
	// iteration will stop
	//
	// The third argument is a channel which may be used for cancelling
	// iteration. Simply send an empty struct value to the channel,
	// and iteration will discontinue. This argument may be nil.
	Each(rc interface{}) (*Iterator, error)

	SetBatchSize(size uint) Query

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

	requestT
}

type queryT struct {
	inst interface{}
	op   opTypeT

	instId    *string
	orderBy   []string
	limit     *int
	skip      *int
	count     *int
	batchSize int
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
	if body, err := defaultClient.doRequest(q); err != nil {
		return err
	} else {
		return handleResponse(body, q.inst)
	}
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
	qv := encodeForRequest(v)
	q.where[f] = qv
	return q
}

func (q *queryT) NotEqualTo(f string, v interface{}) Query {
	qv := encodeForRequest(v)
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

func (q *queryT) MatchesQuery(f string, sq Query) Query {
	q.where[f] = map[string]interface{}{
		"$inQuery": sq,
	}
	return q
}

func (q *queryT) DoesNotMatchQuery(f string, sq Query) Query {
	q.where[f] = map[string]interface{}{
		"$notInQuery": sq,
	}
	return q
}

func (q *queryT) Clone() Query {
	nq := queryT{
		inst:               q.inst,
		op:                 q.op,
		instId:             q.instId,
		currentSession:     q.currentSession,
		shouldUseMasterKey: q.shouldUseMasterKey,
		className:          q.className,
	}

	if q.limit != nil {
		nq.limit = new(int)
		*nq.limit = *q.limit
	}

	if q.skip != nil {
		nq.skip = new(int)
		*nq.skip = *q.skip
	}

	if q.count != nil {
		nq.count = new(int)
		*nq.count = *q.count
	}

	if q.where != nil {
		nq.where = map[string]interface{}{}
		for k, v := range q.where {
			nq.where[k] = v
		}
	}

	if q.include != nil {
		nq.include = map[string]struct{}{}
		for k, v := range q.include {
			nq.include[k] = v
		}
	}

	if q.keys != nil {
		nq.keys = map[string]struct{}{}
		for k, v := range q.keys {
			nq.keys[k] = v
		}
	}

	return &nq
}

func (q *queryT) Sub() Query {
	q2, _ := NewQuery(q.inst)
	return q2
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

var chanInterfaceType = reflect.TypeOf(make(chan interface{}, 0))

func (q *queryT) Each(rc interface{}) (*Iterator, error) {
	instType := reflect.TypeOf(q.inst)
	rv := reflect.ValueOf(rc)
	rt := rv.Type()
	if rt.Kind() != reflect.Chan {
		return nil, fmt.Errorf("rc must be a channel, received %s", rt.Kind())
	}

	if rt.Elem().Kind() == reflect.Ptr {
		if rt.Elem() != instType && rt != chanInterfaceType {
			return nil, fmt.Errorf("1rc must be of type chan %s, received chan %s", instType, rt.Elem())
		}
	} else {
		if rt.Elem() != instType.Elem() && rt != chanInterfaceType {
			return nil, fmt.Errorf("2rc must be of type chan %s, received chan %s", instType.Elem(), rt.Elem())
		}
	}

	if q.op == otInval {
		q.op = otQuery
	}

	if q.limit != nil || q.skip != nil || len(q.orderBy) > 0 {
		return nil, errors.New("cannot iterate over a query with a sort, limit, or skip")
	}

	q.OrderBy("objectId")
	if q.batchSize > 0 {
		q.Limit(q.batchSize)
	} else {
		q.Limit(100)
	}

	i := newIterator()

	go func() {
		defer func() {
			rv.Close()
			close(i.resChan)
			i.iterating = false
		}()

		i.iterating = true

		var sliceType reflect.Type
		if rt == chanInterfaceType {
			sliceType = reflect.SliceOf(instType)
		} else {
			sliceType = reflect.SliceOf(rt.Elem())
		}

		crv := reflect.ValueOf(i.cancel)
		selectCases := []reflect.SelectCase{
			{
				Dir:  reflect.SelectRecv,
				Chan: crv,
			},
			{
				Dir:  reflect.SelectSend,
				Chan: rv,
			},
		}
	loop:
		for {
			select {
			case <-i.cancel:
				break loop
			default:
			}

			s := reflect.New(sliceType)
			s.Elem().Set(reflect.MakeSlice(sliceType, 0, 100))

			// TODO: handle errors and retry if possible
			b, err := defaultClient.doRequest(q)
			if err != nil {
				i.err = err
				i.resChan <- err
				return
			}

			if err := handleResponse(b, s.Interface()); err != nil && err != ErrNoRows {
				i.err = err
				i.resChan <- err
				return
			}

			for i := 0; i < s.Elem().Len(); i++ {
				selectCases[1].Send = s.Elem().Index(i)
				_case, _, _ := reflect.Select(selectCases)
				if _case == 0 {
					break loop
				}
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
		i.resChan <- nil
	}()

	return i, nil
}

func (q *queryT) SetBatchSize(size uint) Query {
	if size <= 1000 {
		q.batchSize = int(size)
	} else {
		q.batchSize = 100
	}
	return q
}

func (q *queryT) Find() error {
	q.op = otQuery
	if b, err := defaultClient.doRequest(q); err != nil {
		return err
	} else {
		return handleResponse(b, q.inst)
	}
}

func (q *queryT) First() error {
	q.op = otQuery
	l := 1
	q.limit = &l

	rv := reflect.ValueOf(q.inst)
	rvi := reflect.Indirect(rv)

	if rvi.Kind() == reflect.Struct {
		dv := reflect.New(reflect.SliceOf(rvi.Type()))
		dv.Elem().Set(reflect.MakeSlice(reflect.SliceOf(rvi.Type()), 0, 1))

		if b, err := defaultClient.doRequest(q); err != nil {
			return err
		} else if err := handleResponse(b, dv.Interface()); err != nil {
			return err
		}

		dvi := reflect.Indirect(dv)
		if dvi.Len() > 0 {
			rv.Elem().Set(dv.Elem().Index(0))
		}
	} else if rvi.Kind() == reflect.Slice {
		if b, err := defaultClient.doRequest(q); err != nil {
			return err
		} else if err := handleResponse(b, q.inst); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("expected struct or slice, got %s", rvi.Kind())
	}
	return nil
}

func (q *queryT) Count() (int64, error) {
	l := 0
	c := 1
	q.limit = &l
	q.count = &c

	var count int64
	if b, err := defaultClient.doRequest(q); err != nil {
		return 0, err
	} else {
		err := handleResponse(b, &count)
		return count, err
	}
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

//
type Iterator struct {
	err       error
	mu        sync.Mutex
	iterating bool
	cancel    chan int
	resChan   chan error
}

func newIterator() *Iterator {
	return &Iterator{
		cancel:  make(chan int, 1),
		resChan: make(chan error, 1),
	}
}

// Returns the terminal error value of the iteration process, or nil if
// the iteration process exited normally (or hasn't started yet)
func (i *Iterator) Error() error {
	return i.err
}

// Cancel interating over the current query. This is a no-op if iteration has
// already terminated
func (i *Iterator) Cancel() {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.iterating {
		i.cancel <- 1
	}
}

// Cancel iterating over the current query, and set the iterator's error value
// to the provided error.
func (i *Iterator) CancelError(err error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.iterating {
		i.err = err
		i.cancel <- 1
	}
}

// Returns a channel that is closed once iteration is finished. Any error causing
// iteration to terminate prematurely will be available on this channel.
func (i *Iterator) Done() <-chan error {
	return i.resChan
}
