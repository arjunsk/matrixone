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
	"gonum.org/v1/gonum/mat"
	"math"
)

// These functions are exposed externally via SQL API.

func Add[T types.RealNumbers](v1, v2 []T) ([]T, error) {
	vec, err := ToGonumVectors[T](v1, v2)
	if err != nil {
		return nil, err
	}

	vec[0].AddVec(vec[0], vec[1])
	return ToMoArray[T](vec[0]), nil
}

func Subtract[T types.RealNumbers](v1, v2 []T) ([]T, error) {
	vec, err := ToGonumVectors[T](v1, v2)
	if err != nil {
		return nil, err
	}

	vec[0].SubVec(vec[0], vec[1])
	return ToMoArray[T](vec[0]), nil
}

func Multiply[T types.RealNumbers](v1, v2 []T) ([]T, error) {
	vec, err := ToGonumVectors[T](v1, v2)
	if err != nil {
		return nil, err
	}

	vec[0].MulElemVec(vec[0], vec[1])
	return ToMoArray[T](vec[0]), nil
}

func Divide[T types.RealNumbers](v1, v2 []T) ([]T, error) {
	// pre-check for division by zero
	for i := 0; i < len(v2); i++ {
		if v2[i] == 0 {
			return nil, moerr.NewDivByZeroNoCtx()
		}
	}

	vec, err := ToGonumVectors[T](v1, v2)
	if err != nil {
		return nil, err
	}

	vec[0].DivElemVec(vec[0], vec[1])
	return ToMoArray[T](vec[0]), nil
}

// Compare returns an integer comparing two arrays/vectors lexicographically.
// TODO: this function might not be correct. we need to compare using tolerance for float values.
// TODO: need to check if we need len(v1)==len(v2) check.
func Compare[T types.RealNumbers](v1, v2 []T) int {
	minLen := len(v1)
	if len(v2) < minLen {
		minLen = len(v2)
	}

	for i := 0; i < minLen; i++ {
		if v1[i] < v2[i] {
			return -1
		} else if v1[i] > v2[i] {
			return 1
		}
	}

	if len(v1) < len(v2) {
		return -1
	} else if len(v1) > len(v2) {
		return 1
	}
	return 0
}

/* ------------ [START] Performance critical functions. ------- */

func InnerProduct[T types.RealNumbers](v1, v2 []T) (float64, error) {

	vec, err := ToGonumVectors[T](v1, v2)
	if err != nil {
		return 0, err
	}

	return mat.Dot(vec[0], vec[1]), nil
}

func L2Distance[T types.RealNumbers](v1, v2 []T) (float64, error) {
	vec, err := ToGonumVectors[T](v1, v2)
	if err != nil {
		return 0, err
	}

	diff := mat.NewVecDense(vec[0].Len(), nil)
	diff.SubVec(vec[0], vec[1])

	return math.Sqrt(mat.Dot(diff, diff)), nil
}

func CosineDistance[T types.RealNumbers](v1, v2 []T) (float64, error) {
	cosineSimilarity, err := CosineSimilarity[T](v1, v2)
	if err != nil {
		return 0, err
	}

	return 1 - cosineSimilarity, nil
}

func CosineSimilarity[T types.RealNumbers](v1, v2 []T) (float64, error) {

	vec, err := ToGonumVectors[T](v1, v2)
	if err != nil {
		return 0, err
	}

	dotProduct := mat.Dot(vec[0], vec[1])

	normVec1 := mat.Norm(vec[0], 2)
	normVec2 := mat.Norm(vec[1], 2)

	if normVec1 == 0 || normVec2 == 0 {
		return 0, moerr.NewInternalErrorNoCtx("cosine_similarity: one of the vectors is zero")
	}

	cosineSimilarity := dotProduct / (normVec1 * normVec2)

	// Handle precision issues. Clamp the cosine_similarity to the range [-1, 1].
	if cosineSimilarity > 1.0 {
		cosineSimilarity = 1.0
	} else if cosineSimilarity < -1.0 {
		cosineSimilarity = -1.0
	}

	// NOTE: Downcast the float64 cosine_similarity to float32 and check if it is
	// 1.0 or -1.0 to avoid precision issue.
	//
	//  Example for corner case:
	// - cosine_similarity(a,a) = 1:
	// - Without downcasting check, we get the following results:
	//   cosine_similarity( [0.46323407, 23.498016, 563.923, 56.076736, 8732.958] ,
	//					    [0.46323407, 23.498016, 563.923, 56.076736, 8732.958] ) =   0.9999999999999998
	// - With downcasting, we get the following results:
	//   cosine_similarity( [0.46323407, 23.498016, 563.923, 56.076736, 8732.958] ,
	//					    [0.46323407, 23.498016, 563.923, 56.076736, 8732.958] ) =   1
	//
	//  Reason:
	// The reason for this check is
	// 1. gonums mat.Dot, mat.Norm returns float64. In other databases, we mostly do float32 operations.
	// 2. float64 operations are not exact.
	// mysql> select 76586261.65813679/(8751.35770370157 *8751.35770370157);
	//+-----------------------------------------------------------+
	//| 76586261.65813679 / (8751.35770370157 * 8751.35770370157) |
	//+-----------------------------------------------------------+
	//|                                            1.000000000000 |
	//+-----------------------------------------------------------+
	//mysql> select cast(76586261.65813679 as double)/(8751.35770370157 * 8751.35770370157);
	//+---------------------------------------------------------------------------+
	//| cast(76586261.65813679 as double) / (8751.35770370157 * 8751.35770370157) |
	//+---------------------------------------------------------------------------+
	//|                                                        0.9999999999999996 |
	//+---------------------------------------------------------------------------+
	// 3. We only need to handle the case for 1.0 and -1.0 with float32 precision.
	//    Rest of the cases can have float64 precision.
	cosineSimilarityF32 := float32(cosineSimilarity)
	if cosineSimilarityF32 == 1 {
		cosineSimilarity = 1
	} else if cosineSimilarityF32 == -1 {
		cosineSimilarity = -1
	}

	return cosineSimilarity, nil
}

func NormalizeL2[T types.RealNumbers](v1 []T) ([]T, error) {

	vec := ToGonumVector[T](v1)

	norm := mat.Norm(vec, 2)
	if norm == 0 {
		return nil, moerr.NewInternalErrorNoCtx("normalize_l2: cannot normalize a zero vector")
	}

	vec.ScaleVec(1/norm, vec)

	return ToMoArray[T](vec), nil
}

// L1Norm returns l1 distance to origin.
func L1Norm[T types.RealNumbers](v []T) (float64, error) {
	vec := ToGonumVector[T](v)

	return mat.Norm(vec, 1), nil
}

// L2Norm returns l2 distance to origin.
func L2Norm[T types.RealNumbers](v []T) (float64, error) {
	vec := ToGonumVector[T](v)

	return mat.Norm(vec, 2), nil
}

/* ------------ [END] Performance critical functions. ------- */

/* ------------ [START] mat.VecDense not supported functions ------- */

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

func Summation[T types.RealNumbers](v []T) (float64, error) {
	n := len(v)
	var sum float64 = 0
	for i := 0; i < n; i++ {
		sum += float64(v[i])
	}
	return sum, nil
}

func Cast[I types.RealNumbers, O types.RealNumbers](in []I) (out []O, err error) {
	n := len(in)

	out = make([]O, n)
	for i := 0; i < n; i++ {
		out[i] = O(in[i])
	}

	return out, nil
}

func VectorSlice[T types.RealNumbers](v []T, start, len int64) ([]T, error) {
	vec := ToGonumVector[T](v)
	res := vec.SliceVec(int(start), int(len))
	return ToMoArray[T](res.(*mat.VecDense)), nil
}

/* ------------ [END] mat.VecDense not supported functions ------- */
