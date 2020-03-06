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

type (
	// ErrorEncoder encode error to response body
	ErrorEncoder func(ctx context.Context, err error) interface{}

	// ResponseEncoder encode payload to response body
	ResponseEncoder func(ctx context.Context, payload interface{}) interface{}

	// fn represents a handler that contains a bundle of hooks
	fn struct {
		plugins []PluginFunc
		adapter adapter
	}
)

var (
	errorEncoder    ErrorEncoder
	responseEncoder ResponseEncoder
)

func failure(ctx context.Context, w http.ResponseWriter, err error) {
	statusCode := http.StatusBadRequest
	if v, ok := UnwrapErrorStatusCode(err); ok {
		statusCode = v
	}
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(errorEncoder(ctx, err))
}

func success(ctx context.Context, w http.ResponseWriter, data interface{}) {
	if data == nil || (reflect.ValueOf(data).Kind() == reflect.Ptr && reflect.ValueOf(data).IsNil()) {
		w.WriteHeader(http.StatusNoContent)
	} else {
		_ = json.NewEncoder(w).Encode(responseEncoder(ctx, data))
	}
}

func (fn *fn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var (
		ctx  = r.Context()
		err  error
		resp interface{}
	)

	for _, b := range globalPlugins {
		ctx, err = b(ctx, r)
		if err != nil {
			failure(ctx, w, err)
			return
		}
	}

	for _, b := range fn.plugins {
		ctx, err = b(ctx, r)
		if err != nil {
			failure(ctx, w, err)
			return
		}
	}

	resp, err = fn.adapter.invoke(ctx, w, r)
	if err != nil {
		failure(ctx, w, err)
		return
	}
	success(ctx, w, resp)
}

func (fn *fn) Plugin(before ...PluginFunc) *fn {
	for _, b := range before {
		if b != nil {
			fn.plugins = append(fn.plugins, b)
		}
	}
	return fn
}

func init() {
	errorEncoder = func(ctx context.Context, err error) interface{} {
		return err.Error()
	}

	responseEncoder = func(ctx context.Context, payload interface{}) interface{} {
		return payload
	}
}
