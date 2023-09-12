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
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestFaissClustering_ComputeCenters(t *testing.T) {
	rowCnt := 3000
	dims := 5
	data := make([][]float32, rowCnt)
	loadData(rowCnt, dims, data)

	clusterCnt := 10
	var cluster Clustering = &faissClustering{}
	centers, err := cluster.ComputeClusters(int64(clusterCnt), data)
	require.Nil(t, err)

	require.Equal(t, 10, len(centers))
}

func loadData(nb int, d int, xb [][]float32) {
	for r := 0; r < nb; r++ {
		xb[r] = make([]float32, d)
		for c := 0; c < d; c++ {
			xb[r][c] = rand.Float32() * 1000
		}
	}
}
