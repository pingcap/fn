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
// See the License for the specific language governing permissions and
// limitations under the License.

package fn

import (
	"reflect"
)

func Wrap(f interface{}) *fn {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		panic("fn only support wrap a function to http.Handler")
	}

	numOut := t.NumOut()

	// Supported signatures
	// func(...) (Response, error)
	if numOut != 2 {
		panic("unsupported function type, function return values should contain response data & error")
	}

	var (
		adapter   adapter
		numIn     = t.NumIn()
		inContext = false
	)

	if numIn > 0 {
		for i := 0; i < numIn; i++ {
			// Legal: func(ctx context.Context, ...) ...
			if t.In(i) == contextType {
				// Illegal: func(..., ctx context.Context, ...) ...
				if i != 0 {
					panic("the `context.Context` must be the first parameter if the signature contains `context.Context`")
				}
				// Illegal: func(..., ctx context.Context, ..., ctx2 context.Context) ...
				if inContext {
					panic("the function can receive two `context.Context`")
				}
				inContext = true
			}
		}
	}

	if numIn == 0 {
		// func() (Response, error)
		adapter = &simplePlainAdapter{
			inContext: false,
			method:    reflect.ValueOf(f),
			cacheArgs: []reflect.Value{},
		}
	} else if numIn == 1 && inContext {
		// func(ctx context.Context) (Response, error)
		adapter = &simplePlainAdapter{
			inContext: true,
			method:    reflect.ValueOf(f),
			cacheArgs: make([]reflect.Value, 1),
		}
	} else if numIn == 1 && !isBuiltinType(t.In(0)) && t.In(0).Kind() == reflect.Ptr {
		// func(request *Customized) (Response, error)
		adapter = &simpleUnaryAdapter{
			argType:   t.In(0),
			method:    reflect.ValueOf(f),
			cacheArgs: make([]reflect.Value, 1),
		}
	} else {
		// Complicated signatures
		//
		// e.g:
		// type LoginResponse {...}
		// type LoginRequest {...}
		//
		// func (header http.Header) (*LoginResponse, error) {}
		// func (form fn.Form) (*LoginResponse, error) {}
		// func (header http.Header, form fn.Form, body io.ReadCloser) (*LoginResponse, error) {}
		// func (header http.Header, r *LoginRequest, url *url.URL) (*LoginResponse, error) { }
		adapter = makeGenericAdapter(reflect.ValueOf(f), inContext)
	}

	return &fn{adapter: adapter}
}

func SetErrorEncoder(c ErrorEncoder) {
	if c == nil {
		panic("nil pointer to error encoder")
	}
	errorEncoder = c
}

func SetResponseEncoder(c ResponseEncoder) {
	if c == nil {
		panic("nil pointer to error encoder")
	}
	responseEncoder = c
}

func SetMultipartFormMaxMemory(m int64) {
	maxMemory = m
}
