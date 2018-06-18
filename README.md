# failure

[![CircleCI](https://circleci.com/gh/morikuni/failure/tree/master.svg?style=shield)](https://circleci.com/gh/morikuni/failure/tree/master)
[![GoDoc](https://godoc.org/github.com/morikuni/failure?status.svg)](https://godoc.org/github.com/morikuni/failure)
[![Go Report Card](https://goreportcard.com/badge/github.com/morikuni/failure)](https://goreportcard.com/report/github.com/morikuni/failure)
[![codecov](https://codecov.io/gh/morikuni/failure/branch/master/graph/badge.svg)](https://codecov.io/gh/morikuni/failure)

failure is a utility package for handling an application error.

The pacakge privides an error below.

```go
// Failure is an error representing failure of something.
type Failure struct {
	// Code is a error code to handle the error in your source code.
	Code Code
	// Message is a error message for the application user.
	// So the message should be human-readable and should be helpful.
	Message string
	// CallStack is a call stack at the time of the error occurs.
	CallStack CallStack
	// Info is information on why the error occurred.
	Info Info
	// Underlying is a underlying error.
	Underlying error
}
```

## Example

```go
package main

import (
	"fmt"

	"github.com/morikuni/failure"
)

const (
	NotFound  failure.Code = "not_found"
	Forbidden failure.Code = "forbidden"
)

func GetX(id int) (int, error) {
	return 0, failure.New(NotFound, "X does not exist.", failure.Info{"id": id})
}

func main() {
	_, err := GetX(123)

	err = failure.Translate(err, Forbidden, "You have no grants to access to X.", failure.Info{"hello": "world"})
	fmt.Println(err)
	// main(forbidden): GetX(not_found)
	fmt.Println(failure.CodeOf(err))
	// forbidden
	fmt.Println(failure.MessageOf(err))
	// You have no grants to access to X.
	fmt.Println(failure.InfosOf(err))
	// [map[hello:world] map[id:123]]
	fmt.Println(failure.CallStackOf(err))
	// [GetX] /Users/morikuni/go/src/github.com/morikuni/failure/example/main.go:15
	// [main] /Users/morikuni/go/src/github.com/morikuni/failure/example/main.go:19
	// [main] /usr/local/Cellar/go/1.10.1/libexec/src/runtime/proc.go:198
	// [goexit] /usr/local/Cellar/go/1.10.1/libexec/src/runtime/asm_amd64.s:2361
}
```
