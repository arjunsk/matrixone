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
	"reflect"
	"testing"
)

func TestBytesToArray(t *testing.T) {
	type args struct {
		input []byte
	}
	type testCase[T any] struct {
		name    string
		args    args
		wantRes []T
	}
	tests := []testCase[float64]{
		{
			name:    "Test1",
			args:    args{input: []byte{0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 8, 64}},
			wantRes: []float64{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := BytesToVector[float64](tt.args.input); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("BytesToVector() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestArrayToBytes(t *testing.T) {
	type args[T any] struct {
		input []T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []byte
	}
	tests := []testCase[float64]{
		{
			name: "Test1",
			args: args[float64]{input: []float64{1, 2, 3}},
			want: []byte{0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 8, 64},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VectorToBytes[float64](tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VectorToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArrayToString(t *testing.T) {
	type args[T any] struct {
		input []T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want string
	}
	tests := []testCase[float64]{
		{
			name: "Test1",
			args: args[float64]{input: []float64{1, 2, 3}},
			want: "[1, 2, 3]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VectorToString[float64](tt.args.input); got != tt.want {
				t.Errorf("VectorToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArraysToString(t *testing.T) {
	type args[T any] struct {
		input [][]T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want string
	}
	tests := []testCase[float64]{
		{
			name: "Test2",
			args: args[float64]{input: [][]float64{{1, 2, 3}, {4, 5, 6}}},
			want: "[1, 2, 3] [4, 5, 6]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VectorsToString[float64](tt.args.input); got != tt.want {
				t.Errorf("VectorsToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToArray(t *testing.T) {
	type args struct {
		input string
	}
	type testCase[T any] struct {
		name string
		args args
		want []T
	}

	float32tests := []testCase[float32]{
		{
			name: "Test3",
			args: args{input: "[1,2,3]"},
			want: []float32{1, 2, 3},
		},
	}
	for _, tt := range float32tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := StringToVector[float32](tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringToVector() = %v, want %v", got, tt.want)
			}
		})
	}

	float64tests := []testCase[float64]{
		{
			name: "Test3",
			args: args{input: "[1,2,3]"},
			want: []float64{1, 2, 3},
		},
	}
	for _, tt := range float64tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := StringToVector[float64](tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringToVector() = %v, want %v", got, tt.want)
			}
		})
	}

}
