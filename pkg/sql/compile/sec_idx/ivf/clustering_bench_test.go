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

package ivf

import (
	"testing"
)

func Benchmark_clustering(b *testing.B) {
	rowCnt := 1_000
	dims := 1024
	data := make([][]float32, rowCnt)
	loadData(rowCnt, dims, data)

	faiss := NewFaissClustering()
	kmeans := NewKmeansClustering()

	b.Run("FAISS", func(b *testing.B) {
		b.ResetTimer()
		for i := 1; i < b.N; i++ {
			_, err := faiss.ComputeClusters(int64(i), data)
			if err != nil {
				panic(err)
			}
		}
	})

	b.Run("KMEANS", func(b *testing.B) {
		b.ResetTimer()
		for i := 1; i < b.N; i++ {
			_, err := kmeans.ComputeClusters(int64(i), data)
			if err != nil {
				panic(err)
			}
		}
	})

	/*
		rowCnt := 1_000
		dims := 1024
		Benchmark_clustering/FAISS-10         	     100	 144196691 ns/op
		Benchmark_clustering/KMEANS-10        	     100	3705940079 ns/op
	*/
}
