# barefn

This library aims to simplify the construction of JSON API service,
`barefn.Wrap` is able to wrap any function to adapt the interface of
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
barefn.Form        // request.Form
barefn.PostForm    // request.PostForm
*barefn.Form       // request.Form
*barefn.PostForm   // request.PostForm
*url.URL           // request.URL
*multipart.Form    // request.MultipartForm
*http.Request      // raw request
```

## Usage

```
http.Handle("/test", barefn.Wrap(test))

func test(io.ReadCloser, http.Header, barefn.Form, barefn.PostForm, *CustomizedRequestType, *url.URL, *multipart.Form) (*CustomizedResponseType, error)
```

## Examples

```go
package examples

import (
    "io"
    "mime/multipart"
	"net/http"
    "net/url"
	
	"github.com/pingcap/barefn"
)

type Request struct{
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct{
	Token string `json:"token"`
}

func api1() (*Response, error) {
	return &Response{ Token: "token" }, nil
}

func api2(request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{ Token: token }, nil
}

func api3(rawreq *http.Request, request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{ Token: token }, nil
}

func api4(rawreq http.Header, request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{ Token: token }, nil
}

func api5(form *barefn.Form, request *Request) (*Response, error) {
	token := request.Username + request.Password + form.Get("type")
	return &Response{ Token: token }, nil
}

func api6(body io.ReadCloser, request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{ Token: token }, nil
}

func api7(form *multipart.Form, request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{ Token: token }, nil
}

func api7(urls *url.URL, request *Request) (*Response, error) {
	token := request.Username + request.Password
	return &Response{ Token: token }, nil
}
```