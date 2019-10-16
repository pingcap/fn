# fn

This library aims to simplify the construction of JSON API service,
`fn.Wrap` is able to wrap any function to adapt the interface of
`http.Handler`, which unmarshals POST data to a struct automatically.

## Benchmark

```
BenchmarkIsBuiltinType-8                50000000                33.5 ns/op             0 B/op          0 allocs/op
BenchmarkSimplePlainAdapter_Invoke-8     2000000               757 ns/op             195 B/op          3 allocs/op
BenchmarkSimpleUnaryAdapter_Invoke-8     2000000               681 ns/op             946 B/op          5 allocs/op
BenchmarkGenericAdapter_Invoke-8         2000000               708 ns/op             946 B/op          5 allocs/op
```

## Support types

```
io.ReadCloser      // request.Body
http.Header        // request.Header
fn.Form        // request.Form
fn.PostForm    // request.PostForm
*fn.Form       // request.Form
*fn.PostForm   // request.PostForm
*url.URL           // request.URL
*multipart.Form    // request.MultipartForm
*http.Request      // raw request
```

## Usage

```go
http.Handle("/test", fn.Wrap(test))

func test(io.ReadCloser, http.Header, fn.Form, fn.PostForm, *CustomizedRequestType, *url.URL, *multipart.Form) (*CustomizedResponseType, error)
```

## Examples

### Basic

```go
package examples

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/pingcap/fn"
)

type Request struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Token string `json:"token"`
}

func api1() (*Response, error) {
	return &Response{Token: "token"}, nil
}

func api2(request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{Token: token}, nil
}

func api3(rawreq *http.Request, request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{Token: token}, nil
}

func api4(rawreq http.Header, request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{Token: token}, nil
}

func api5(form *fn.Form, request *Request) (*Response, error) {
	token := request.Username + request.Password + form.Get("type")
	return &Response{Token: token}, nil
}

func api6(body io.ReadCloser, request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{Token: token}, nil
}

func api7(form *multipart.Form, request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{Token: token}, nil
}

func api7(urls *url.URL, request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{Token: token}, nil
}

func api8(urls *url.URL, form *multipart.Form, body io.ReadCloser, rawreq http.Header, request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{Token: token}, nil
}
```

### Plugins

```go
package examples

import (
	"context"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/pingcap/fn"
)

var PermissionDenied = errors.New("permission denied")

func logger(ctx context.Context, req *http.Request) (context.Context, error) {
	log.Println("Request", req.RemoteAddr, req.URL.String())
	return ctx, nil
}

func ipWhitelist(ctx context.Context, req *http.Request) (context.Context, error) {
	if strings.HasPrefix(req.RemoteAddr, "172.168") {
		return ctx, PermissionDenied
	}
	return ctx, nil
}

func auth(ctx context.Context, req *http.Request) (context.Context, error) {
	token := req.Header.Get("X-Auth-token")
	_ = token // Validate token (e.g: query db)
	if token != "valid" {
		return ctx, fn.ErrorWithStatusCode(PermissionDenied, http.StatusForbidden)
	}
	return ctx, nil
}

type Request struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Token string `json:"token"`
}

func example() {
	fn.Plugin(logger, ipWhitelist, auth)
	http.Handle("/api1", fn.Wrap(api1))
	http.Handle("/api2", fn.Wrap(api2))
}

// api1 and api2 request have be validated by `ipWhitelist` and `auth`

func api1() (*Response, error) {
	return &Response{Token: "token"}, nil
}

func api2(request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{Token: token}, nil
}
```

### `fn.Group`

```go
package examples

import (
	"context"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/pingcap/fn"
)

var PermissionDenied = errors.New("permission denied")

func logger(ctx context.Context, req *http.Request) (context.Context, error) {
	log.Println("Request", req.RemoteAddr, req.URL.String())
	return ctx, nil
}

func ipWhitelist(ctx context.Context, req *http.Request) (context.Context, error) {
	if strings.HasPrefix(req.RemoteAddr, "172.168") {
		return ctx, PermissionDenied
	}
	return ctx, nil
}

func auth(ctx context.Context, req *http.Request) (context.Context, error) {
	token := req.Header.Get("X-Auth-token")
	_ = token // Validate token (e.g: query db)
	if token != "valid" {
		return ctx, fn.ErrorWithStatusCode(PermissionDenied, http.StatusForbidden)
	}
	return ctx, nil
}

type User struct {
	Balance int64
}

func queryUserFromRedis(ctx context.Context, req *http.Request) (context.Context, error) {
	token := req.Header.Get("X-Auth-token")
	_ = token // Validate token (e.g: query db)
	if token != "valid" {
		return ctx, fn.ErrorWithStatusCode(PermissionDenied, http.StatusForbidden)
	}
	user := &User{
		Balance: 10000, // balance from redis
	}
	return context.WithValue(ctx, "user", user), nil
}

type Response struct {
	Balance int64 `json:"balance"`
}

func example() {
	// Global plugins
	fn.Plugin(logger, ipWhitelist, auth)

	group := fn.NewGroup()

	// Group plugins
	group.Plugin(queryUserFromRedis)
	http.Handle("/user/balance", group.Wrap(fetchBalance))
	http.Handle("/user/buy", group.Wrap(buy))
}

func fetchBalance(ctx context.Context) (*Response, error) {
	user := ctx.Value("user").(*User)
	return &Response{Balance: user.Balance}, nil
}

func buy(ctx context.Context) (*Response, error) {
	user := ctx.Value("user").(*User)
	if user.Balance < 100 {
		return nil, errors.New("please check balance")
	}
	user.Balance -= 100
	return &Response{Balance: user.Balance}, nil
}
```

### ResponseEncoder

```go
package examples

import (
	"context"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/pingcap/fn"
)

var PermissionDenied = errors.New("permission denied")

func logger(ctx context.Context, req *http.Request) (context.Context, error) {
	log.Println("Request", req.RemoteAddr, req.URL.String())
	return ctx, nil
}

func ipWhitelist(ctx context.Context, req *http.Request) (context.Context, error) {
	if strings.HasPrefix(req.RemoteAddr, "172.168") {
		return ctx, PermissionDenied
	}
	return ctx, nil
}

func auth(ctx context.Context, req *http.Request) (context.Context, error) {
	token := req.Header.Get("X-Auth-token")
	_ = token // Validate token (e.g: query db)
	if token != "valid" {
		return ctx, fn.ErrorWithStatusCode(PermissionDenied, http.StatusForbidden)
	}
	return ctx, nil
}

func injectRequest(ctx context.Context, req *http.Request) (context.Context, error) {
	return context.WithValue(ctx, "_rawreq", req), nil
}

type User struct {
	Balance int64
}

func queryUserFromRedis(ctx context.Context, req *http.Request) (context.Context, error) {
	token := req.Header.Get("X-Auth-token")
	_ = token // Validate token (e.g: query db)
	if token != "valid" {
		return ctx, fn.ErrorWithStatusCode(PermissionDenied, http.StatusForbidden)
	}
	user := &User{
		Balance: 10000, // balance from redis
	}
	return context.WithValue(ctx, "user", user), nil
}

type Response struct {
	Balance int64 `json:"balance"`
}

type ResponseMessage struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

type ErrorMessage struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func example() {
	// Global plugins
	fn.Plugin(logger, ipWhitelist, auth, injectRequest)
	// Uniform all responses
	fn.SetErrorEncoder(func(ctx context.Context, err error) interface{} {
		req := ctx.Value("_rawreq").(*http.Request)
		log.Println("Error occurred: ", req.URL, err)
		return &ErrorMessage{
			Code:  -1,
			Error: err.Error(),
		}
	})

	fn.SetResponseEncoder(func(ctx context.Context, payload interface{}) interface{} {
		return &ResponseMessage{
			Code: 1,
			Data: payload,
		}
	})

	group := fn.NewGroup()

	// Group plugins
	group.Plugin(queryUserFromRedis)
	http.Handle("/user/balance", group.Wrap(fetchBalance))
	http.Handle("/user/buy", group.Wrap(buy))
}

func fetchBalance(ctx context.Context) (*Response, error) {
	user := ctx.Value("user").(*User)
	return &Response{Balance: user.Balance}, nil
}

func buy(ctx context.Context) (*Response, error) {
	user := ctx.Value("user").(*User)
	if user.Balance < 100 {
		return nil, errors.New("please check balance")
	}
	user.Balance -= 100
	return &Response{Balance: user.Balance}, nil
}
```