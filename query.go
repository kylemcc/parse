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
	otCreate
	otUpdate
	otDelete
)

func (o opTypeT) method() string {
	switch o {
	case otGet, otQuery:
		return "GET"
	case otCreate:
		return "POST"
	case otUpdate:
		return "PUT"
	case otDelete:
		return "DELETE"
	default:
		return "GET"
	}
}

type Query interface {
	UseMasterKey() Query

	Get(id string) error

	OrderBy(f string) Query
	OrderByFields(fs ...string) Query
	Limit(l int) Query
	Skip(s int) Query
	Include(f string) Query
	Keys(fs ...string) Query

	EqualTo(f string, v interface{}) Query
	NotEqualTo(f string, v interface{}) Query
	GreaterThan(f string, v interface{}) Query
	GreaterThanOrEqual(f string, v interface{}) Query
	LessThan(f string, v interface{}) Query
	LessThanOrEqual(f string, v interface{}) Query
	In(f string, vs ...interface{}) Query
	NotIn(f string, vs ...interface{}) Query
	Exists(f string) Query
	DoesNotExist(f string) Query
	All(f string, vs ...interface{}) Query
	Contains(f string, v string) Query
	StartsWith(f string, v string) Query
	EndsWith(f string, v string) Query
	Or(qs ...Query) Query

	Each(rc interface{}, ec chan<- error) error
	Find() error
	First() error
	Count() (int64, error)
}

type queryT struct {
	inst interface{}
	op   opTypeT

	instId  *string
	orderBy []string
	limit   *int
	skip    *int
	count   *int
	where   map[string]interface{}
	include map[string]struct{}
	keys    map[string]struct{}

	currentSession *sessionT

	shouldUseMasterKey bool
}

func NewQuery(v interface{}) (Query, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, errors.New("v must be a non-nil pointer")
	}

	return &queryT{
		inst:    v,
		orderBy: make([]string, 0),
		where:   make(map[string]interface{}),
		include: make(map[string]struct{}),
		keys:    make(map[string]struct{}),
	}, nil
}

func (q *queryT) UseMasterKey() Query {
	q.shouldUseMasterKey = true
	return q
}

// Get retrieves the instance of the type pointed to by v and
// identified by id, and stores the result in v.
//
// v should should be a pointer to a struct represting the type
// to be retrieved.
func (q *queryT) Get(id string) error {
	q.op = otGet
	q.instId = &id
	return defaultClient.doRequest(q, q.inst)
}

func (q *queryT) OrderBy(f string) Query {
	q.orderBy = []string{f}
	return q
}

func (q *queryT) OrderByFields(fs ...string) Query {
	q.orderBy = append(make([]string, len(fs)), fs...)
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

func (q *queryT) Include(f string) Query {
	q.include[f] = struct{}{}
	return q
}

func (q *queryT) Keys(fs ...string) Query {
	for _, f := range fs {
		q.include[f] = struct{}{}
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
	return nil
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
		is := make([]string, len(q.include))
		for k := range q.include {
			is = append(is, k)
		}
		i := strings.Join(is, ",")
		p["include"] = []string{i}
	}

	if len(q.keys) > 0 {
		ks := make([]string, len(q.include))
		for k := range q.keys {
			ks = append(ks, k)
		}
		k := strings.Join(ks, ",")
		p["include"] = []string{k}
	}

	return p.Encode(), nil
}

// Implement the operationT interface
func (q *queryT) method() string {
	return q.op.method()
}

func (q *queryT) endpoint() (string, error) {
	u := url.URL{}
	var p string

	var inst interface{}

	rt := reflect.TypeOf(q.inst)
	rt = rt.Elem()
	if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
		rte := rt.Elem()
		var rv reflect.Value
		if rte.Kind() == reflect.Ptr {
			rv = reflect.New(rte.Elem())
		} else {
			rv = reflect.New(rte)
		}
		inst = rv.Interface()
	} else {
		inst = q.inst
	}

	if v, ok := inst.(iParseEp); ok {
		p = v.Endpoint()
	} else {
		var cname string
		if v, ok := inst.(iClassName); ok {
			cname = v.ClassName()
		} else {
			t := reflect.TypeOf(inst)
			cname = t.Elem().Name()
		}
		p = path.Join("classes", cname)
	}

	p = path.Join(ParseVersion, p)

	switch q.op {
	case otGet, otUpdate, otDelete:
		p = path.Join(p, *q.instId)
	}

	qs, err := q.payload()
	if err != nil {
		return "", err
	}

	u.Scheme = "https"
	u.Host = "api.parse.com"
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

// From the Javascript library - convert the string represented by re into a regex
// value that matches it. MongoDb (what backs Parse) uses PCRE syntax
func quote(re string) string {
	return "\\Q" + strings.Replace(re, "\\E", "\\E\\\\E\\Q", -1) + "\\E"
}

func getQueryRepr(inst interface{}, f string, v interface{}) interface{} {
	fmt.Printf("inst:[%+v] f:[%s] v:[%v]\n", inst, f, v)
	var fname string
	fieldMap := getFieldNameMap(reflect.ValueOf(inst))
	fmt.Printf("fnamemap:[%+v]\n", fieldMap)

	if fn, ok := fieldMap[f]; ok {
		fname = fn
	} else {
		fname = f
	}
	fname = strings.Title(fname)

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
						return map[string]string{
							"__type": "Date",
							"iso": v.(string),
						}
					}
				} else {
					var id string
					var cname string
					fv := rviInst.FieldByName(fname)
					fvi := reflect.Indirect(fv)

					if tmp, ok := fv.Interface().(iClassName); ok {
						cname = tmp.ClassName()
					} else {
						cname = fvi.Type().Name()
					}

					rv := reflect.ValueOf(v)
					rvi := reflect.Indirect(rv)
					if rvi.Kind() == reflect.String {
						id = rvi.Interface().(string)
					} else if idf := rvi.FieldByName("Id"); idf.IsValid() {
						id = idf.Interface().(string)
					}

					return Pointer{
						Id: id,
						ClassName: cname,
					}
				}
			}
		}
	}

	return v
}
