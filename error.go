// Copyright 2020 PingCAP, Inc.
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

type statusCodeError struct {
	err        error
	statusCode int
}

type StatusCodeError interface {
	StatusCode() int
}

func (s *statusCodeError) StatusCode() int {
	return s.statusCode
}

func (s *statusCodeError) Unwrap() error {
	return s.err
}

func (s *statusCodeError) Error() string {
	return s.err.Error()
}

func UnwrapErrorStatusCode(err error) (int, bool) {
	for err != nil {
		if v, ok := err.(StatusCodeError); ok {
			return v.StatusCode(), true
		}
		err = Unwrap(err)
	}
	return 0, false
}

func ErrorWithStatusCode(err error, statusCode int) error {
	return &statusCodeError{err, statusCode}
}
