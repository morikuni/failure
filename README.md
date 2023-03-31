# failure

[![CircleCI](https://circleci.com/gh/morikuni/failure/tree/main.svg?style=shield)](https://circleci.com/gh/morikuni/failure/tree/master)
[![Go Reference](https://pkg.go.dev/badge/github.com/morikuni/failure.svg)](https://pkg.go.dev/github.com/morikuni/failure)
[![Go Report Card](https://goreportcard.com/badge/github.com/morikuni/failure)](https://goreportcard.com/report/github.com/morikuni/failure)
[![codecov](https://codecov.io/gh/morikuni/failure/branch/main/graph/badge.svg)](https://codecov.io/gh/morikuni/failure)

Package `failure` provides errors utilities for your application errors.

- Automatically generate awesome `err.Error` message for developers.
- Flexible error messages for end users.
- Powerful and readable stack trace.
- Error context, such as function parameter, with key-value data.
- Extensible error chain.

## Usage

At first, define error codes for your application.

```go
const (
	NotFound failure.StringCode = "NotFound"
	InvalidArgument failure.StringCode = "InvalidArgument"
	Internal failure.StringCode = "Internal"
)
```

Using `failure.New`, return an error with error code.

```go
return failure.New(NotFound)
```

Handle the error with `failure.Is` and translate it into another error code with `failure.Translate`.

```go
if failure.Is(err, NotFound) {
	return failure.Translate(err, Internal)
}
```

If you want to just return the error, use `failure.Wrap`.

```go
if err != nil {
	return failure.Wrap(err)
}
```

An error context and message for end user can be attached.

```go
func Foo(a, b string) error {
	return failure.New(InvalidArgument, 
		failure.Context{"a": a, "b": b},
		failure.Message("Given parameters are invalid!!"),
	)
}
```

Awesome error message for developers.

```go
func main() {
	err := Bar()
	fmt.Println(err)
	fmt.Println("=====")
	fmt.Printf("%+v\n", err)
}

func Bar() error {
	err := Foo("hello", "world")
	if err != nil {
		return failure.Wrap(err)
	}
	return nil
}
```

```
main.Bar: main.Foo: a=hello b=world: Given parameters are invalid!!: code(InvalidArgument)
=====
[main.Bar] /tmp/sandbox615088634/prog.go:25
[main.Foo] /tmp/sandbox615088634/prog.go:31
    a = hello
    b = world
    message("Given parameters are invalid!!")
    code(InvalidArgument)
[CallStack]
    [main.Foo] /tmp/sandbox615088634/prog.go:31
    [main.Bar] /tmp/sandbox615088634/prog.go:23
    [main.main] /tmp/sandbox615088634/prog.go:16
    [runtime.main] /usr/local/go-faketime/src/runtime/proc.go:204
    [runtime.goexit] /usr/local/go-faketime/src/runtime/asm_amd64.s:1374
```

`package.FunctionName` like `main.Bar` and `main.Foo` is automatically added to error message.
With `%+v` format, it prints the detailed error chain + the call stack of the first error.

## Full Example for HTTP Server

Try it on [The Go Playground](https://play.golang.org/p/Pmgm7-7J1_c)!
	
```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"

	"github.com/morikuni/failure"
)

// error codes for your application.
const (
	NotFound  failure.StringCode = "NotFound"
	Forbidden failure.StringCode = "Forbidden"
)

func GetACL(projectID, userID string) (acl interface{}, e error) {
	notFound := true
	if notFound {
		return nil, failure.New(NotFound,
			failure.Context{"project_id": projectID, "user_id": userID},
		)
	}
	return nil, failure.Unexpected("unexpected error")
}

func GetProject(projectID, userID string) (project interface{}, e error) {
	_, err := GetACL(projectID, userID)
	if err != nil {
		if failure.Is(err, NotFound) {
			return nil, failure.Translate(err, Forbidden,
				failure.Message("no acl exists"),
				failure.Context{"additional_info": "hello"},
			)
		}
		return nil, failure.Wrap(err)
	}
	return nil, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	_, err := GetProject(r.FormValue("project_id"), r.FormValue("user_id"))
	if err != nil {
		HandleError(w, err)
		return
	}
}

func getHTTPStatus(err error) int {
	c, ok := failure.CodeOf(err)
	if !ok {
		return http.StatusInternalServerError
	}
	switch c {
	case NotFound:
		return http.StatusNotFound
	case Forbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func getMessage(err error) string {
	msg, ok := failure.MessageOf(err)
	if ok {
		return msg
	}
	return "Error"
}

func HandleError(w http.ResponseWriter, err error) {
	w.WriteHeader(getHTTPStatus(err))
	io.WriteString(w, getMessage(err))

	fmt.Println("============ Error ============")
	fmt.Printf("Error = %v\n", err)

	code, _ := failure.CodeOf(err)
	fmt.Printf("Code = %v\n", code)

	msg, _ := failure.MessageOf(err)
	fmt.Printf("Message = %v\n", msg)

	cs, _ := failure.CallStackOf(err)
	fmt.Printf("CallStack = %v\n", cs)

	fmt.Printf("Cause = %v\n", failure.CauseOf(err))

	fmt.Println()
	fmt.Println("============ Detail ============")
	fmt.Printf("%+v\n", err)
	// [main.GetProject] /go/src/github.com/morikuni/failure/example/main.go:36
	//     message("no acl exists")
	//     additional_info = hello
	//     code(Forbidden)
	// [main.GetACL] /go/src/github.com/morikuni/failure/example/main.go:21
	//     project_id = 123
	//     user_id = 456
	//     code(NotFound)
	// [CallStack]
	//     [main.GetACL] /go/src/github.com/morikuni/failure/example/main.go:21
	//     [main.GetProject] /go/src/github.com/morikuni/failure/example/main.go:33
	//     [main.Handler] /go/src/github.com/morikuni/failure/example/main.go:47
	//     [http.HandlerFunc.ServeHTTP] /usr/local/go/src/net/http/server.go:1964
	//     [http.(*ServeMux).ServeHTTP] /usr/local/go/src/net/http/server.go:2361
	//     [http.serverHandler.ServeHTTP] /usr/local/go/src/net/http/server.go:2741
	//     [http.(*conn).serve] /usr/local/go/src/net/http/server.go:1847
	//     [runtime.goexit] /usr/local/go/src/runtime/asm_amd64.s:1333
}

func main() {
	req := httptest.NewRequest(http.MethodGet, "/?project_id=aaa&user_id=111", nil)
	rec := httptest.NewRecorder()
	Handler(rec, req)

	res, _ := httputil.DumpResponse(rec.Result(), true)
	fmt.Println("============ Dump ============")
	fmt.Println(string(res))
}
```

</details>
