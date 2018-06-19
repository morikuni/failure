# failure

[![CircleCI](https://circleci.com/gh/morikuni/failure/tree/master.svg?style=shield)](https://circleci.com/gh/morikuni/failure/tree/master)
[![GoDoc](https://godoc.org/github.com/morikuni/failure?status.svg)](https://godoc.org/github.com/morikuni/failure)
[![Go Report Card](https://goreportcard.com/badge/github.com/morikuni/failure)](https://goreportcard.com/report/github.com/morikuni/failure)
[![codecov](https://codecov.io/gh/morikuni/failure/branch/master/graph/badge.svg)](https://codecov.io/gh/morikuni/failure)

failure is a utility package for handling application errors.
It is inspired by [https://middlemost.com/failure-is-your-domain](https://middlemost.com/failure-is-your-domain) and [github.com/pkg/errors](https://github.com/pkg/errors).

The package provides an error type below.

```go
// Failure is an error representing failure of something.
type Failure struct {
	// Code is an error code to handle the error in your source code.
	Code Code
	// Message is an error message for the application user.
	// So the message should be human-readable and be helpful.
	Message string
	// CallStack is a call stack at the time of the error occurred.
	CallStack CallStack
	// Info is optional information on why the error occurred.
	Info Info
	// Underlying is an underlying error of the failure.
	Underlying error
}
```

## Example

The failure works with error codes you defined in your application.

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/morikuni/failure"
)

// error codes for your application.
const (
	NotFound  failure.Code = "not_found"
	Forbidden failure.Code = "forbidden"
)

func GetACL(projectID, userID string) (acl interface{}, e error) {
	notFound := true
	if notFound {
		return nil, failure.New(NotFound).WithInfo(failure.Info{"projectID": projectID, "userID": userID})
	}
	err := errors.New("error")
	if err != nil {
		return nil, failure.Wrap(err)
	}
	return nil, nil
}

func GetProject(projectID, userID string) (project interface{}, e error) {
	_, err := GetACL(projectID, userID)
	if err != nil {
		switch failure.CodeOf(err) {
		case NotFound:
			return nil, failure.Translate(err, Forbidden).
				WithMessage("You have no grant to access the project.").
				WithInfo(failure.Info{"additionalInfo": "hello"})
		default:
			return nil, err
		}
	}
	return nil, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	_, err := GetProject(r.FormValue("projectID"), r.FormValue("userID"))
	if err != nil {
		HandleError(w, err)
		return
	}
}

func HandleError(w http.ResponseWriter, err error) {
	switch failure.CodeOf(err) {
	case NotFound:
		w.WriteHeader(http.StatusNotFound)
	case Forbidden:
		w.WriteHeader(http.StatusForbidden)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	io.WriteString(w, failure.MessageOf(err))

	// The failure contains useful messages for creating response and debugging.

	fmt.Println(err)
	// GetProject(forbidden): GetACL(not_found)
	fmt.Printf("%+v\n", err)
	// GetProject(forbidden): GetACL(not_found)
	//   [GetACL] /go/src/github.com/morikuni/failure/example/main.go:21
	//   [GetProject] /go/src/github.com/morikuni/failure/example/main.go:31
	//   [Handler] /go/src/github.com/morikuni/failure/example/main.go:46
	//   [HandlerFunc.ServeHTTP] /usr/local/go/src/net/http/server.go:1918
	//   [(*ServeMux).ServeHTTP] /usr/local/go/src/net/http/server.go:2254
	//   [serverHandler.ServeHTTP] /usr/local/go/src/net/http/server.go:2619
	//   [(*conn).serve] /usr/local/go/src/net/http/server.go:1801
	//   [goexit] /usr/local/go/src/runtime/asm_amd64.s:2337
	fmt.Println(failure.CodeOf(err))
	// forbidden
	fmt.Println(failure.MessageOf(err))
	// You have no grant to access the project.
	fmt.Println(failure.InfoListOf(err))
	// [map[additionalInfo:hello] map[projectID:123 userID:456]]
	fmt.Println(failure.CallStackOf(err))
	// GetACL: GetProject: Handler: HandlerFunc.ServeHTTP: (*ServeMux).ServeHTTP: serverHandler.ServeHTTP: (*conn).serve: goexit
}

func main() {
	http.HandleFunc("/", Handler)
	http.ListenAndServe(":8080", nil)
	// $ http "localhost:8080/?projectID=123&userID=456"
	// HTTP/1.1 403 Forbidden
	// Content-Length: 40
	// Content-Type: text/plain; charset=utf-8
	// Date: Tue, 19 Jun 2018 02:56:32 GMT
	//
	// You have no grant to access the project.
}
```
