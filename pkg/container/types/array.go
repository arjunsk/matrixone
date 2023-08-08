// Copyright 2023 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"bytes"
	"fmt"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"io"
	"strconv"
	"strings"
	"unsafe"
)

const (
	//MaxArrayDimension Comment: https://github.com/arjunsk/matrixone/pull/35#discussion_r1275713689
	MaxArrayDimension = 65536
)

func BytesToArray[T RealNumbers](input []byte) (res []T) {
	return DecodeSlice[T](input)
}

func ArrayToBytes[T RealNumbers](input []T) []byte {
	return EncodeSlice[T](input)
}

func ArrayToString[T RealNumbers](input []T) string {
	var buffer bytes.Buffer
	_, _ = io.WriteString(&buffer, "[")
	for i, value := range input {
		if i > 0 {
			_, _ = io.WriteString(&buffer, ", ")
		}
		_, _ = io.WriteString(&buffer, fmt.Sprintf("%v", value))
	}
	_, _ = io.WriteString(&buffer, "]")
	return buffer.String()
}

func ArraysToString[T RealNumbers](input [][]T) string {
	strValues := make([]string, len(input))
	for i, row := range input {
		strValues[i] = ArrayToString[T](row)
	}
	return strings.Join(strValues, " ")
}

func StringToArray[T RealNumbers](input string) ([]T, error) {
	input = strings.ReplaceAll(input, "[", "")
	input = strings.ReplaceAll(input, "]", "")
	input = strings.ReplaceAll(input, " ", "")

	numStrs := strings.Split(input, ",")
	result := make([]T, len(numStrs))

	var t T
	for i, numStr := range numStrs {
		switch any(t).(type) {

		case int8:
			num, err := strconv.ParseInt(numStr, 10, 8)
			if err != nil {
				return nil, moerr.NewInternalErrorNoCtx("Error while parsing array : %v", err)
			}
			numi8 := int8(num)
			result[i] = *(*T)(unsafe.Pointer(&numi8))
		case int16:
			num, err := strconv.ParseInt(numStr, 10, 16)
			if err != nil {
				return nil, moerr.NewInternalErrorNoCtx("Error while parsing array : %v", err)
			}
			numi16 := int16(num)
			result[i] = *(*T)(unsafe.Pointer(&numi16))
		case int32:
			num, err := strconv.ParseInt(numStr, 10, 32)
			if err != nil {
				return nil, moerr.NewInternalErrorNoCtx("Error while parsing array : %v", err)
			}
			numi32 := int32(num)
			result[i] = *(*T)(unsafe.Pointer(&numi32))
		case int64:
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				return nil, moerr.NewInternalErrorNoCtx("Error while parsing array : %v", err)
			}
			result[i] = *(*T)(unsafe.Pointer(&num))

		case float32:
			num, err := strconv.ParseFloat(numStr, 32)
			if err != nil {
				return nil, moerr.NewInternalErrorNoCtx("Error while parsing array : %v", err)
			}
			// FIX: https://stackoverflow.com/a/36391858/1609570
			numf32 := float32(num)
			result[i] = *(*T)(unsafe.Pointer(&numf32))
		case float64:
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return nil, moerr.NewInternalErrorNoCtx("Error while parsing array : %v", err)
			}
			result[i] = *(*T)(unsafe.Pointer(&num))
		default:
			panic(moerr.NewInternalErrorNoCtx("not implemented"))
		}

	}

	return result, nil
}

func CompareArray[T RealNumbers](left, right []T) int {

	if len(left) != len(right) {
		//TODO: check this with Min.
		panic(moerr.NewInternalErrorNoCtx("Dimensions should be same"))
	}

	for i := 0; i < len(left); i++ {
		if left[i] == right[i] {
			continue
		} else if left[i] > right[i] {
			return +1
		} else {
			return -1
		}
	}

	return 0
}
