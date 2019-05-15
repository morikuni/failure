# failure

[![CircleCI](https://circleci.com/gh/morikuni/failure/tree/master.svg?style=shield)](https://circleci.com/gh/morikuni/failure/tree/master)
[![GoDoc](https://godoc.org/github.com/morikuni/failure?status.svg)](https://godoc.org/github.com/morikuni/failure)
[![Go Report Card](https://goreportcard.com/badge/github.com/morikuni/failure)](https://goreportcard.com/report/github.com/morikuni/failure)
[![codecov](https://codecov.io/gh/morikuni/failure/branch/master/graph/badge.svg)](https://codecov.io/gh/morikuni/failure)

Package failure provides an error represented as error code and extensible error interface with wrappers.

Use error code instead of error type.

```go
var NotFound failure.StringCode = "NotFound"

err := failure.New(NotFound)

if failure.Is(err, NotFound) { // true
	r.WriteHeader(http.StatusNotFound)
}
```

Wrap errors.

```go
type Wrapper interface {
	WrapError(err error) error
}

err := failure.Wrap(err, MarkTemporary())
```

Unwrap errors with Iterator.

```go
type Unwrapper interface {
	UnwrapError() error
}

i := failure.NewIterator(err)
for i.Next() { // unwrap error
	err := i.Error()
	if e, ok := err.(Temporary); ok {
		return e.IsTemporary()
	}
}
```

## Example

### You can try it on [The Go Playground](https://play.golang.org/p/Pmgm7-7J1_c)

You can also see the example on GitHub by opening this.

<details>
	
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
