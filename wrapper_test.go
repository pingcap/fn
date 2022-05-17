// Copyright 2017 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	. "github.com/pingcap/check"
)

type fnSuite struct{}

func TestFn(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&fnSuite{})

type testRequest struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}
type testResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type testErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"message"`
}

var successResponse = &testResponse{Message: "success"}

func init() {
	Plugin(
		func(ctx context.Context, request *http.Request) (context.Context, error) {
			return context.WithValue(ctx, "global1", "globalvalue1"), nil
		},
		nil,
		func(ctx context.Context, request *http.Request) (context.Context, error) {
			return context.WithValue(ctx, "global2", "globalvalue2"), nil
		})
}

// acceptable function signature
func withNone() (*testResponse, error)                         { return successResponse, nil }
func withBody(io.ReadCloser) (*testResponse, error)            { return successResponse, nil }
func withReq(*testRequest) (*testResponse, error)              { return successResponse, nil }
func withHeader(http.Header) (*testResponse, error)            { return successResponse, nil }
func withForm(Form) (*testResponse, error)                     { return successResponse, nil }
func withPostForm(PostForm) (*testResponse, error)             { return successResponse, nil }
func withFormPtr(*Form) (*testResponse, error)                 { return successResponse, nil }
func withPostFormPtr(*PostForm) (*testResponse, error)         { return successResponse, nil }
func withMultipartForm(*multipart.Form) (*testResponse, error) { return successResponse, nil }
func withUrl(*url.URL) (*testResponse, error)                  { return successResponse, nil }
func withRawRequest(*http.Request) (*testResponse, error)      { return successResponse, nil }

func withInContext(context.Context) (*testResponse, error) { return successResponse, nil }

func withInContextAndPayload(context.Context, *testRequest) (*testResponse, error) {
	return successResponse, nil
}

func withMulti(*testRequest, Form, PostForm, http.Header, *url.URL) (*testResponse, error) {
	return nil, nil
}
func withAll(io.ReadCloser, *testRequest, Form, PostForm, http.Header, *multipart.Form, *url.URL) (*testResponse, error) {
	return nil, nil
}

func (s *fnSuite) TestHandler(c *C) {
	Wrap(withNone)
	Wrap(withBody)
	Wrap(withReq)
	Wrap(withHeader)
	Wrap(withForm)
	Wrap(withPostForm)
	Wrap(withFormPtr)
	Wrap(withPostFormPtr)
	Wrap(withMultipartForm)
	Wrap(withUrl)
	Wrap(withRawRequest)
	Wrap(withMulti)
	Wrap(withAll)
	Wrap(withInContext)
	Wrap(withInContextAndPayload)
}

func (s *fnSuite) TestPlugin(c *C) {
	logic := func(ctx context.Context) (*testResponse, error) {
		c.Assert(ctx.Value("key").(string) == "value", IsTrue)
		c.Assert(ctx.Value("key2").(string) == "value2", IsTrue)
		return &testResponse{}, nil
	}

	plugin1 := func(ctx context.Context, request *http.Request) (context.Context, error) {
		c.Assert(ctx.Value("global1").(string) == "globalvalue1", IsTrue)
		c.Assert(ctx.Value("global2").(string) == "globalvalue2", IsTrue)
		return context.WithValue(ctx, "key", "value"), nil
	}

	plugin2 := func(ctx context.Context, request *http.Request) (context.Context, error) {
		c.Assert(ctx.Value("global1").(string) == "globalvalue1", IsTrue)
		c.Assert(ctx.Value("global2").(string) == "globalvalue2", IsTrue)
		c.Assert(ctx.Value("key").(string) == "value", IsTrue)
		return context.WithValue(ctx, "key2", "value2"), nil
	}

	handler := Wrap(logic).Plugin(plugin1, plugin2)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "", nil)
	c.Assert(err, IsNil)
	handler.ServeHTTP(recorder, request)
}

func (s *fnSuite) TestGroupPlugin(c *C) {
	group := NewGroup()
	group.Plugin(func(ctx context.Context, request *http.Request) (context.Context, error) {
		return context.WithValue(ctx, "key", "value"), nil
	})

	logic := func(ctx context.Context) (*testResponse, error) {
		c.Assert(ctx.Value("key").(string) == "value", IsTrue)
		return &testResponse{}, nil
	}
	handler := group.Wrap(logic)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "", nil)
	c.Assert(err, IsNil)
	handler.ServeHTTP(recorder, request)
}

func (s *fnSuite) TestSetResponseEncoder(c *C) {
	handler := Wrap(func(ctx context.Context, request *http.Request) (context.Context, error) {
		return nil, nil
	})

	testResp := &testResponse{
		Code:    1,
		Message: "msg",
	}
	SetResponseEncoder(func(ctx context.Context, payload interface{}) interface{} {
		return testResp
	})

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "", nil)
	c.Assert(err, IsNil)
	handler.ServeHTTP(recorder, request)
	respMsg := &testResponse{}
	_ = json.Unmarshal(recorder.Body.Bytes(), &respMsg)
	c.Assert(reflect.DeepEqual(respMsg, testResp), IsTrue)
}

func (s *fnSuite) TestSetErrorEncoder(c *C) {
	handler := Wrap(func(ctx context.Context, request *http.Request) (context.Context, error) {
		return nil, errors.New("")
	})

	testErrorResp := &testErrorResponse{
		Code:  -1,
		Error: "something went wrong",
	}
	SetErrorEncoder(func(ctx context.Context, err error) interface{} {
		return testErrorResp
	})

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "", nil)
	c.Assert(err, IsNil)
	handler.ServeHTTP(recorder, request)

	respMsg := &testErrorResponse{}
	_ = json.Unmarshal(recorder.Body.Bytes(), &respMsg)
	c.Assert(reflect.DeepEqual(respMsg, testErrorResp), IsTrue)
}

func (s *fnSuite) TestGenericAdapter_Invoke(c *C) {
	type CustomForm testRequest
	handler := Wrap(func(ctx context.Context, form *CustomForm) (context.Context, error) {
		return nil, nil
	})

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "", nil)
	c.Assert(err == nil, IsTrue)
	payload := []byte(`{"for":"hello", "bar":10000}`)
	request.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
	c.Assert(err, IsNil)
	handler.ServeHTTP(recorder, request)
}

func (s *fnSuite) TestSimpleUnaryAdapter_Invoke(c *C) {
	handler := Wrap(withReq)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		c.Fatal(err)
	}
	payload := []byte(`{"for":"hello", "bar":10000}`)
	request.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
	c.Assert(err, IsNil)
	handler.ServeHTTP(recorder, request)
}

func (s *fnSuite) TestErrorWithStatusCode(c *C) {
	handler := Wrap(func(ctx context.Context, request *http.Request) (context.Context, error) {
		return nil, ErrorWithStatusCode(errors.New("not found"), http.StatusNotFound)
	})

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "", nil)
	c.Assert(err, IsNil)
	handler.ServeHTTP(recorder, request)
	c.Assert(recorder.Code == http.StatusNotFound, IsTrue)
}

func BenchmarkSimplePlainAdapter_Invoke(b *testing.B) {
	handler := Wrap(withNone)
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		b.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(recorder, request)
	}
}

func BenchmarkSimpleUnaryAdapter_Invoke(b *testing.B) {
	handler := Wrap(withReq)
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		b.Fatal(err)
	}
	payload := []byte(`{"for":"hello", "bar":10000}`)
	request.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
	recorder := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(recorder, request)
	}
}

func BenchmarkGenericAdapter_Invoke(b *testing.B) {
	handler := Wrap(withMulti)
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		b.Fatal(err)
	}
	payload := []byte(`{"for":"hello", "bar":10000}`)
	request.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
	recorder := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(recorder, request)
	}
}
