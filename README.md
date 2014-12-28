#Parse

[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/kylemcc/parse) [![license](http://img.shields.io/badge/license-BSD-red.svg?style=flat)](https://raw.githubusercontent.com/kylemcc/parse/master/LICENSE)

This package provides a client for Parse's REST API. So far, it supports most of the query operations
provided by Parse's [Javascript library](https://parse.com/docs/js/symbols/Parse.Query.html), with a
few exceptions (listed below under TODO).

###Installation

    go get github.com/kylemcc/parse

###Documentation
[Full documentation](http://godoc.org/github.com/kylemcc/parse) is provided by [godoc.org](http://godoc.org)

###Usage:
```go
package main

import (
    "fmt"
	"time"
    
    "github.com/kylemcc/parse"
)

func main() {
    parse.Initialize("APP_ID", "REST_KEY", "MASTER_KEY") // master key is optional
    
    user := parse.User{}
    q, err := parse.NewQuery(&user)
	if err != nil {
		panic(err)
	}
    q.EqualTo("email", "kylemcc@gmail.com")
    q.GreaterThan("numFollowers", 10).OrderBy("-createdAt") // API is chainable
    err := q.First()
    if err != nil {
        if pe, ok := err.(parse.ParseError); ok {
            fmt.Printf("Error querying parse: %d - %s\n", pe.Code(), pe.Message())
        }
    }
    
    fmt.Printf("Retrieved user with id: %s\n", u.Id)

	q2, _ := parse.NewQuery(&parse.User{})
	q2.GreaterThan("createdAt", time.Date(2014, 01, 01, 0, 0, 0, 0, time.UTC))
	rc := make(chan *parse.User)
	ec := make(chan error)

	// .Each will retrieve all results for a query and send them to the provided channel
	q2.Each(rc, ec)
	for {
		select {
		case u, ok := <-rc:
			if ok {
				fmt.Printf("received user: %v\n", u)
			} else {
				rc = nil
			}
		case err, ok := <-ec:
			if ok {
				fmt.Printf("error: %v\n", err)
			} else {
				ec = nil
			}
		}
		if rc == nil && ec == nil {
			break
		}
	}
}
```

###TODO
- Missing query operations
	- Match query ($inQuery)
	- Related to
- Missing CRUD operations:
    - Update
		- Field ops (__op):
			- AddRelation
			- RemoveRelation
- ACLs
- Cloud Functions
- Background Jobs
- Push Notifications
- Config
- Analytics
- File upload/retrieval
