// Copyright 2020-present, lizc2003@gmail.com
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

package util

import (
	"fmt"
)

const (
	ErrFail = 1
)

func NewErrorWithCode(code int32, text string) error {
	return &errorWithCode{c: code, s: text}
}

func GetErrorCode(err error) int32 {
	if err == nil {
		return 0
	}

	if e, ok := err.(*errorWithCode); ok {
		return e.Code()
	} else {
		return ErrFail
	}
}

type errorWithCode struct {
	c int32
	s string
}

func (e *errorWithCode) Error() string {
	return fmt.Sprintf("error: code = %d, desc = %s", e.c, e.s)
}

func (e *errorWithCode) Code() int32 {
	return e.c
}
