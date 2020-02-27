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
	"errors"
	. "github.com/pingcap/check"
	"net/http"
	"testing"
)

var (
	errTest = errors.New("test")
)

type withError struct {
	err error
}

func (w *withError) Error() string {
	return w.err.Error()
}

func (w *withError) Unwrap() error {
	return w.err
}

type errSuite struct{}

func TestError(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&errSuite{})

// TestWithError test UnwrapErrorStatusCode method
func (e *errSuite) TestWithError(c *C) {
	var err error = &withError{
		err: ErrorWithStatusCode(errTest, http.StatusInternalServerError),
	}
	code, ok := UnwrapErrorStatusCode(err)
	c.Assert(ok, IsTrue)
	c.Assert(code == http.StatusInternalServerError, IsTrue)
}

// TestOriginError ErrorWithStatusCode Unwrap
func (e *errSuite) TestOriginError(c *C) {
	err := ErrorWithStatusCode(errTest, http.StatusInternalServerError)
	err = Unwrap(err)
	c.Assert(err == errTest, IsTrue)
}
