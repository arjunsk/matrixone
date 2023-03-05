// Copyright 2021 - 2022 Matrix Origin
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

package inside

import (
	"github.com/matrixorigin/matrixone/pkg/container/nulls"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

type typeFunc func(typ types.Type) (bool, int32)

// InternalCharSize Implementation of mo internal function 'internal_char_size'
func InternalCharSize(vectors []*vector.Vector, proc *process.Process) (*vector.Vector, error) {
	return generalInternalType("internal_char_size", vectors, proc, getTypeCharSize)
}

// InternalCharLength Implementation of mo internal function 'internal_char_length'
func InternalCharLength(vectors []*vector.Vector, proc *process.Process) (*vector.Vector, error) {
	return generalInternalType("internal_char_length", vectors, proc, getTypeCharLength)
}

// InternalNumericPrecision Implementation of mo internal function 'internal_numeric_precision'
func InternalNumericPrecision(vectors []*vector.Vector, proc *process.Process) (*vector.Vector, error) {
	return generalInternalType("internal_numeric_precision", vectors, proc, getTypeNumericPrecision)
}

// InternalNumericScale Implementation of mo internal function 'internal_numeric_scale'
func InternalNumericScale(vectors []*vector.Vector, proc *process.Process) (*vector.Vector, error) {
	return generalInternalType("internal_numeric_scale", vectors, proc, getTypeNumericScale)
}

// InternalDatetimePrecision Implementation of mo internal function 'internal_datetime_precision'
func InternalDatetimePrecision(vectors []*vector.Vector, proc *process.Process) (*vector.Vector, error) {
	return generalInternalType("internal_datetime_precision", vectors, proc, getTypeDatetimePrecision)
}

// InternalColumnCharacterSet  Implementation of mo internal function 'internal_column_character_set'
func InternalColumnCharacterSet(vectors []*vector.Vector, proc *process.Process) (*vector.Vector, error) {
	return generalInternalType("internal_column_character_set", vectors, proc, getTypeCharacterSet)
}

// Mo General function for obtaining type information
func generalInternalType(funName string, vectors []*vector.Vector, proc *process.Process, typefunc typeFunc) (*vector.Vector, error) {
	inputVector := vectors[0]
	resultType := types.T_int64.ToType()
	inputValues := vector.MustStrCols(inputVector)
	if inputVector.IsScalar() {
		if inputVector.IsScalarNull() {
			return proc.AllocScalarNullVector(resultType), nil
		} else {
			var typ types.Type
			bytes := []byte(inputValues[0])
			err := types.Decode(bytes, &typ)
			if err != nil {
				return nil, err
			}
			isVaild, val := typefunc(typ)
			if isVaild {
				return vector.NewConstFixed(resultType, inputVector.Length(), int64(val), proc.Mp()), nil
			} else {
				return proc.AllocScalarNullVector(resultType), nil
			}
		}
	} else {
		resVector, err := proc.AllocVectorOfRows(resultType, int64(len(inputValues)), inputVector.Nsp)
		if err != nil {
			return nil, err
		}
		resultValues := vector.MustTCols[int64](resVector)
		for i, typeValue := range inputValues {
			var typ types.Type
			bytes := []byte(typeValue)
			err = types.Decode(bytes, &typ)
			if err != nil {
				return nil, err
			}
			isVaild, val := typefunc(typ)
			if isVaild {
				resultValues[i] = int64(val)
			} else {
				nulls.Add(resVector.Nsp, uint64(i))
			}
		}
		return resVector, nil
	}
}

// 'internal_char_lengh' function operator
func getTypeCharLength(typ types.Type) (bool, int32) {
	if types.IsString(typ.Oid) {
		return true, typ.Width
	} else {
		return false, -1
	}
}

// 'internal_char_size' function operator
func getTypeCharSize(typ types.Type) (bool, int32) {
	if types.IsString(typ.Oid) {
		return true, typ.Size * typ.Width
	} else {
		return false, -1
	}
}

// 'internal_numeric_precision' function operator
func getTypeNumericPrecision(typ types.Type) (bool, int32) {
	if types.IsDecimal(typ.Oid) {
		return true, typ.Precision
	} else {
		return false, -1
	}
}

// 'internal_numeric_scale' function operator
func getTypeNumericScale(typ types.Type) (bool, int32) {
	if types.IsDecimal(typ.Oid) {
		return true, typ.Scale
	} else {
		return false, -1
	}
}

// 'internal_datetime_precision' function operator
func getTypeDatetimePrecision(typ types.Type) (bool, int32) {
	if typ.Oid == types.T_datetime {
		return true, typ.Precision
	} else {
		return false, -1
	}
}

// 'internal_column_character_set_name' function operator
func getTypeCharacterSet(typ types.Type) (bool, int32) {
	if typ.Oid == types.T_varchar ||
		typ.Oid == types.T_char ||
		typ.Oid == types.T_blob ||
		typ.Oid == types.T_text {
		return true, int32(typ.Charset)
	} else {
		return false, -1
	}
}
