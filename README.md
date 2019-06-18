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