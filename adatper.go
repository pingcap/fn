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
	"context"
	"encoding/json"
	"net/http"
	"reflect"
)

// adapter represents a container that contain a handler function
// and convert a it to a http.Handler
type adapter interface {
	invoke(context.Context, http.ResponseWriter, *http.Request) (interface{}, error)
}

// genericAdapter represents a common adapter
type genericAdapter struct {
	inContext bool
	method    reflect.Value
	numIn     int
	types     []reflect.Type
	cacheArgs []reflect.Value // cache args
}

// Accept zero parameter adapter
type simplePlainAdapter struct {
	inContext bool
	method    reflect.Value
	cacheArgs []reflect.Value
}

// Accept only one parameter adapter
type simpleUnaryAdapter struct {
	outContext bool
	argType    reflect.Type
	method     reflect.Value
	cacheArgs  []reflect.Value // cache args
}

func makeGenericAdapter(method reflect.Value, inContext bool) *genericAdapter {
	var noSupportExists = false
	t := method.Type()
	numIn := t.NumIn()

	a := &genericAdapter{
		inContext: inContext,
		method:    method,
		numIn:     numIn,
		types:     make([]reflect.Type, numIn),
		cacheArgs: make([]reflect.Value, numIn),
	}

	for i := 0; i < numIn; i++ {
		in := t.In(i)
		if in != contextType && !isBuiltinType(in) {
			if noSupportExists {
				panic("function should accept only one customize type")
			}

			if in.Kind() != reflect.Ptr {
				panic("customize type should be a pointer(" + in.PkgPath() + "." + in.Name() + ")")
			}
			noSupportExists = true
		}
		a.types[i] = in
	}

	return a
}

func (a *genericAdapter) invoke(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	values := a.cacheArgs
	for i := 0; i < a.numIn; i++ {
		typ := a.types[i]
		v, ok := supportTypes[typ]
		if ok {
			value, err := v(r)
			if err != nil {
				return nil, err
			}
			values[i] = value
		} else if typ == contextType {
			values[i] = reflect.ValueOf(ctx)
		} else {
			d := reflect.New(a.types[i].Elem()).Interface()
			err := json.NewDecoder(r.Body).Decode(d)
			if err != nil {
				return nil, err
			}
			values[i] = reflect.ValueOf(d)
		}
	}

	var err error
	results := a.method.Call(values)
	payload := results[0].Interface()
	if e := results[1].Interface(); e != nil {
		err = e.(error)
	}
	return payload, err
}

func (a *simplePlainAdapter) invoke(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	if a.inContext {
		a.cacheArgs[0] = reflect.ValueOf(ctx)
	}

	var err error
	results := a.method.Call(a.cacheArgs)
	payload := results[0].Interface()
	if e := results[1].Interface(); e != nil {
		err = e.(error)
	}
	return payload, err
}

func (a *simpleUnaryAdapter) invoke(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	data := reflect.New(a.argType.Elem()).Interface()
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		return nil, err
	}

	a.cacheArgs[0] = reflect.ValueOf(data)
	results := a.method.Call(a.cacheArgs)
	payload := results[0].Interface()
	if e := results[1].Interface(); e != nil {
		err = e.(error)
	}
	return payload, err
}
