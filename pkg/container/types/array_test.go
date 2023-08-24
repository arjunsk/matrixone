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
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/common/util"
	"reflect"
	"testing"
)

func TestBytesToArray(t *testing.T) {
	type args struct {
		input []byte
	}
	type testCase struct {
		name       string
		args       args
		wantResF32 []float32
		wantResF64 []float64
	}
	tests := []testCase{
		{
			name:       "Test1 - float32",
			args:       args{input: []byte{0, 0, 128, 63, 0, 0, 0, 64, 0, 0, 64, 64}},
			wantResF32: []float32{1, 2, 3},
		},
		{
			name:       "Test2 - float64",
			args:       args{input: []byte{0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 8, 64}},
			wantResF64: []float64{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantResF32 != nil {
				if gotRes := BytesToArray[float32](tt.args.input); !reflect.DeepEqual(gotRes, tt.wantResF32) {
					t.Errorf("BytesToArray() = %v, want %v", gotRes, tt.wantResF32)
				}
			}
			if tt.wantResF64 != nil {
				if gotRes := BytesToArray[float64](tt.args.input); !reflect.DeepEqual(gotRes, tt.wantResF64) {
					t.Errorf("BytesToArray() = %v, want %v", gotRes, tt.wantResF64)
				}
			}
		})
	}
}

func TestArrayToBytes(t *testing.T) {

	type testCase struct {
		name    string
		argsF32 []float32
		argsF64 []float64
		want    []byte
	}
	tests := []testCase{
		{
			name:    "Test1 - Float32",
			argsF32: []float32{1, 2, 3},
			want:    []byte{0, 0, 128, 63, 0, 0, 0, 64, 0, 0, 64, 64},
		},
		{
			name:    "Test2 - Float64",
			argsF64: []float64{1, 2, 3},
			want:    []byte{0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 8, 64},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.argsF32 != nil {
				if got := ArrayToBytes[float32](tt.argsF32); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ArrayToBytes() = %v, want %v", got, tt.want)
				}
			}

			if tt.argsF64 != nil {
				if got := ArrayToBytes[float64](tt.argsF64); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ArrayToBytes() = %v, want %v", got, tt.want)
				}
			}

		})
	}
}

func TestArrayToString(t *testing.T) {
	type testCase struct {
		name    string
		argsF32 []float32
		argsF64 []float64
		want    string
	}
	tests := []testCase{
		{
			name:    "Test1 - Float32",
			argsF32: []float32{1, 2, 3, 4},
			want:    "[1, 2, 3, 4]",
		},
		{
			name:    "Test2 - Float64",
			argsF64: []float64{1, 2, 3},
			want:    "[1, 2, 3]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.argsF32 != nil {
				if got := ArrayToString[float32](tt.argsF32); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ArrayToString() = %v, want %v", got, tt.want)
				}
			}

			if tt.argsF64 != nil {
				if got := ArrayToString[float64](tt.argsF64); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ArrayToString() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestArraysToString(t *testing.T) {
	type testCase struct {
		name    string
		argsF32 [][]float32
		argsF64 [][]float64
		want    string
	}
	tests := []testCase{
		{
			name:    "Test1 - Float32",
			argsF32: [][]float32{{1, 2, 3, 4}, {0, 0, 0, 0}},
			want:    "[1, 2, 3, 4] [0, 0, 0, 0]",
		},
		{
			name:    "Test2 - Float64",
			argsF64: [][]float64{{1, -2, 3, 4}, {0, 0, 0, 0}},
			want:    "[1, -2, 3, 4] [0, 0, 0, 0]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.argsF32 != nil {
				if got := ArraysToString[float32](tt.argsF32); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ArraysToString() = %v, want %v", got, tt.want)
				}
			}

			if tt.argsF64 != nil {
				if got := ArraysToString[float64](tt.argsF64); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ArraysToString() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestStringToArray(t *testing.T) {
	type args struct {
		input string
		typ   T
	}
	type testCase struct {
		name       string
		args       args
		wantResF32 []float32
		wantResF64 []float64
		wantErr    error
	}
	tests := []testCase{
		{
			name:       "Test1 - float32",
			args:       args{input: "[1,2,3,-2]", typ: T_array_float32},
			wantResF32: []float32{1, 2, 3, -2},
		},
		{
			name:       "Test2 - float64",
			args:       args{input: "[1,2,3,30]", typ: T_array_float64},
			wantResF64: []float64{1, 2, 3, 30},
		},
		{
			name:    "Test3 - float64",
			args:    args{input: "[1,2,3,", typ: T_array_float64},
			wantErr: moerr.NewInternalErrorNoCtx("malformed vector input: [1,2,3,"),
		},
		{
			name:    "Test4 - float64",
			args:    args{input: "[]", typ: T_array_float64},
			wantErr: moerr.NewInternalErrorNoCtx("malformed vector input: []"),
		},
		{
			name:    "Test4 - float64",
			args:    args{input: "[1,a]", typ: T_array_float64},
			wantErr: moerr.NewInternalErrorNoCtx("error while casting a to DOUBLE"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.wantResF32 != nil {
				if gotRes, err := StringToArray[float32](tt.args.input); err != nil || !reflect.DeepEqual(gotRes, tt.wantResF32) {
					t.Errorf("StringToArray() = %v, want %v", gotRes, tt.wantResF32)
				}
			}
			if tt.wantResF64 != nil {
				if gotRes, err := StringToArray[float64](tt.args.input); err != nil || !reflect.DeepEqual(gotRes, tt.wantResF64) {
					t.Errorf("StringToArray() = %v, want %v", gotRes, tt.wantResF64)
				}
			}

			if tt.wantErr != nil && tt.args.typ == T_array_float32 {
				if _, gotErr := StringToArray[float32](tt.args.input); gotErr == nil {
					t.Errorf("StringToArray() = %v, want %v", gotErr, tt.wantErr)
				} else {
					if !reflect.DeepEqual(gotErr, tt.wantErr) {
						t.Errorf("StringToArray() = %v, want %v", gotErr, tt.wantErr)
					}
				}
			}

			if tt.wantErr != nil && tt.args.typ == T_array_float64 {
				if _, gotErr := StringToArray[float64](tt.args.input); gotErr == nil {
					t.Errorf("StringToArray() = %v, want %v", gotErr, tt.wantErr)
				} else {
					if !reflect.DeepEqual(gotErr, tt.wantErr) {
						t.Errorf("StringToArray() = %v, want %v", gotErr, tt.wantErr)
					}
				}
			}

		})
	}
}

func TestHexToArray(t *testing.T) {
	type args struct {
		input []byte
	}
	type testCase[T RealNumbers] struct {
		name    string
		args    args
		want    []T
		wantErr bool
	}
	tests := []testCase[float32]{
		{
			name:    "T1",
			args:    args{input: util.UnsafeStringToBytes("7e98b23e9e10383b2f41133f")},
			want:    []float32{0.34881967306137085, 0.0028086076490581036, 0.5752133727073669},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HexToArray[float32](tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("HexToArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HexToArray() got = %v, want %v", got, tt.want)
			}
		})
	}
}
