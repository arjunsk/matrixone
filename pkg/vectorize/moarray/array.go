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

package moarray

import (
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/vectorize/momath"
)

//TODO: Check on optimization.
// 1. Should we return []T or *[]T
// 2. Should we accept v1 *[]T. v1 is a Slice, so I think, it should be pass by reference.
// 3. Later on, use tensor library to improve the performance (may be via GPU)

func Add[T types.RealNumbers](v1, v2 []T) ([]T, error) {
	if len(v1) != len(v2) {
		return nil, moerr.NewArrayInvalidOpNoCtx(len(v1), len(v2))
	}
	n := len(v1)
	r := make([]T, n)
	for i := 0; i < n; i++ {
		r[i] = v1[i] + v2[i]
	}
	return r, nil
}

func Subtract[T types.RealNumbers](v1, v2 []T) ([]T, error) {
	if len(v1) != len(v2) {
		return nil, moerr.NewArrayInvalidOpNoCtx(len(v1), len(v2))
	}
	n := len(v1)
	r := make([]T, n)
	for i := 0; i < n; i++ {
		r[i] = v1[i] - v2[i]
	}
	return r, nil
}

func Multiply[T types.RealNumbers](v1, v2 []T) ([]T, error) {
	if len(v1) != len(v2) {
		return nil, moerr.NewArrayInvalidOpNoCtx(len(v1), len(v2))
	}
	n := len(v1)
	r := make([]T, n)
	for i := 0; i < n; i++ {
		r[i] = v1[i] * v2[i]
	}
	return r, nil
}

func Divide[T types.RealNumbers](v1, v2 []T) ([]T, error) {
	if len(v1) != len(v2) {
		return nil, moerr.NewArrayInvalidOpNoCtx(len(v1), len(v2))
	}
	n := len(v1)
	r := make([]T, n)
	for i := 0; i < n; i++ {
		if v2[i] == 0 {
			return nil, moerr.NewDivByZeroNoCtx()
		}
		r[i] = v1[i] / v2[i]
	}
	return r, nil
}

// Compare the l2_norm between 2 ARRAY's. This is more accurate than element wise comparison because
//  1. there won't be dimension mismatch issue
//  2. for element-wise comparison, the float32[i] value in ARRAY/VECTOR is always going
//     to have some subtle difference between v1 and v2, resulting in full element wise comparison of v1 and v2.
//  3. l2_norm comparison helps in ordering ARRAYs by nearness on the cartesian plane.
func Compare[T types.RealNumbers](v1, v2 []T) int {
	a, _ := L2Norm[T](v1) // you can ignore the l2_norm error.
	b, _ := L2Norm[T](v2)

	if a == b {
		return 0
	}
	if a < b {
		return -1
	}
	return +1
}

func Cast[I types.RealNumbers, O types.RealNumbers](in []I) (out []O) {
	n := len(in)

	out = make([]O, n)
	for i := 0; i < n; i++ {
		out[i] = O(in[i])
	}

	return out
}

func Abs[T types.RealNumbers](v []T) (res []T, err error) {
	n := len(v)
	res = make([]T, n)
	for i := 0; i < n; i++ {
		res[i], err = momath.AbsSigned[T](v[i])
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func Sqrt[T types.RealNumbers](v []T) (res []float64, err error) {
	n := len(v)
	res = make([]float64, n)
	for i := 0; i < n; i++ {
		res[i], err = momath.Sqrt(float64(v[i]))
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func Summation[T types.RealNumbers](v []T) float64 {
	n := len(v)
	var sum float64 = 0
	for i := 0; i < n; i++ {
		sum = sum + float64(v[i])
	}
	return sum
}

func InnerProduct[T types.RealNumbers](v1, v2 []T) (float64, error) {

	if len(v1) != len(v2) {
		return 0, moerr.NewArrayInvalidOpNoCtx(len(v1), len(v2))
	}

	n := len(v1)

	var productVal T
	var sum float64 = 0

	for i := 0; i < n; i++ {
		productVal = v1[i] * v2[i]
		sum = sum + float64(productVal)
	}
	return sum, nil
}

// L1Norm returns l1 distance to origin.
// The only time, this could throw error is when T = int8 (v[i] is -128)
func L1Norm[T types.RealNumbers](v []T) (float64, error) {
	n := len(v)

	var absVal T
	var err error
	var sum float64 = 0

	for i := 0; i < n; i++ {
		absVal, err = momath.AbsSigned[T](v[i])
		if err != nil {
			return 0, err
		}

		sum = sum + float64(absVal)
	}
	return sum, nil
}

// L2Norm returns l2 distance to origin.
// You can ignore the error as math.sqrt will not throw -ve error since,
// sum is always +ve due to sum(pow(v,2)).
func L2Norm[T types.RealNumbers](v []T) (float64, error) {
	n := len(v)

	var sqrVal T
	var sum float64 = 0

	for i := 0; i < n; i++ {
		sqrVal = v[i] * v[i]
		sum = sum + float64(sqrVal)
	}

	return momath.Sqrt(sum)
}

func CosineSimilarity[T types.RealNumbers](v1, v2 []T) (float32, error) {

	if len(v1) != len(v2) {
		return 0, moerr.NewArrayInvalidOpNoCtx(len(v1), len(v2))
	}

	a, err := InnerProduct[T](v1, v2)
	if err != nil {
		return 0, err
	}

	b, err := L2Norm[T](v1)
	if err != nil {
		return 0, err
	}

	c, err := L2Norm[T](v2)
	if err != nil {
		return 0, err
	}

	sum := a / (b * c)
	return float32(sum), nil
}
