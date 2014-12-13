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
    
    "github.com/kylemcc/parse"
)

func main() {
    parse.Initialize("APP_ID", "REST_KEY", "MASTER_KEY") // master key is optional
    
    user := parse.User{}
    q := parse.NewQuery(&user)
    q.EqualTo("email", "kylemcc@gmail.com")
    q.GreaterThan("numFollowers", 10).Limit(1) // API is chainable
    err := q.First()
    if err != nil {
        if pe, ok := err.(parse.ParseError); ok {
            fmt.Printf("Error querying parse: %d - %s\n", pe.Code(), pe.Message())
        }
    }
    
    fmt.Printf("Retrieved user with id: %s\n", u.Id)
}
```

###TODO
- Documentation
- Missing query operations
    - Regex match
    - Match key in subquery
	- First
- Missing CRUD operations:
    - Create
    - Update
    - Delete
- Login/Sessions
- ACLs
- Cloud Functions
- Background Jobs
- Push Notifications
- Config
- Analytics
- File upload/retrieval
