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

package barefn

import (
	"encoding/json"
	"net/http"
)

type (
	// ErrorEncoder encode error to response body
	ErrorEncoder func(error) interface{}

	// ResponseEncoder encode payload to response body
	ResponseEncoder func(payload interface{}) interface{}

	// BareFn represents a handler that contains a bundle of hooks
	BareFn struct {
		plugins []PluginFunc
		adapter adapter
	}

	statusCodeError struct {
		error
		statusCode int
	}

	StatusCodeError interface {
		StatusCode() int
	}
)

var (
	errorEncoder    ErrorEncoder
	responseEncoder ResponseEncoder
)

func (s *statusCodeError) StatusCode() int {
	return s.statusCode
}

func failure(w http.ResponseWriter, err error) {
	statusCode := http.StatusBadRequest
	if v, ok := err.(StatusCodeError); ok {
		statusCode = v.StatusCode()
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorEncoder(err))
}

func success(w http.ResponseWriter, data interface{}) {
	json.NewEncoder(w).Encode(responseEncoder(data))
}

func (fn *BareFn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var (
		ctx  = r.Context()
		err  error
		resp interface{}
	)

	for _, b := range globalPlugins {
		ctx, err = b(ctx, r)
		if err != nil {
			failure(w, err)
			return
		}
	}

	for _, b := range fn.plugins {
		ctx, err = b(ctx, r)
		if err != nil {
			failure(w, err)
			return
		}
	}

	resp, err = fn.adapter.invoke(ctx, w, r)
	if err != nil {
		failure(w, err)
		return
	}
	success(w, resp)
}

func (fn *BareFn) Plugin(before ...PluginFunc) *BareFn {
	for _, b := range before {
		if b != nil {
			fn.plugins = append(fn.plugins, b)
		}
	}
	return fn
}

func init() {
	errorEncoder = func(err error) interface{} {
		return err.Error()
	}

	responseEncoder = func(payload interface{}) interface{} {
		return payload
	}
}
