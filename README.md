# Parse

[![Build Status](https://travis-ci.org/kylemcc/parse.svg?branch=master)](https://travis-ci.org/kylemcc/parse) [![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/kylemcc/parse) [![license](http://img.shields.io/badge/license-BSD-red.svg?style=flat)](https://raw.githubusercontent.com/kylemcc/parse/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/kylemcc/parse)](https://goreportcard.com/report/kylemcc/parse)

This package provides a client for Parse's REST API. So far, it supports most of the query operations
provided by Parse's [Javascript library](https://docs.parseplatform.org/js/guide/), with a
few exceptions (listed below under TODO).

### Installation

    go get github.com/kylemcc/parse

### Documentation
[Full documentation](http://godoc.org/github.com/kylemcc/parse) is provided by [godoc.org](http://godoc.org)

### Usage:
```go
package main

import (
	"fmt"
	"time"

	"github.com/kylemcc/parse"
)

func main() {
	parse.Initialize("APP_ID", "REST_KEY", "MASTER_KEY") // rest and master keys are optional
	parse.ServerURL("https://SERVER_URL/PATH")
	resp, err := parse.ServerHealthCheck()
	if err != nil {
		panic(err)
	}
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

	fmt.Printf("Retrieved user with id: %s\n", user.Id)

	q2, _ := parse.NewQuery(&parse.User{})
	q2.GreaterThan("createdAt", time.Date(2014, 01, 01, 0, 0, 0, 0, time.UTC))

	rc := make(chan *parse.User)

	// .Each will retrieve all results for a query and send them to the provided channel
	// The iterator returned allows for early cancelation of the iteration process, and
	// stores any error that triggers early termination
	iterator, err := q2.Each(rc)
	for u := range rc {
		fmt.Printf("received user: %v\n", u)
		// Do something
		if err := process(u); err != nil {
			// Cancel if there was an error
			iterator.Cancel()
		}
	}

	// An error occurred - not all rows were processed
	if it.Error() != nil {
		panic(it.Error())
	}
}
```

### TODO
- Missing query operations
	- Related to
- Missing CRUD operations:
    - Update
		- Field ops (__op):
			- AddRelation
			- RemoveRelation
- Roles
- Background Jobs
- Analytics
- File upload/retrieval
- Batch operations
