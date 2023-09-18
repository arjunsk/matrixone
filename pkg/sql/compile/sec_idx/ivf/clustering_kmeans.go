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
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
)

type kmeansClustering struct {
	km kmeans.Kmeans
}

func NewKmeansClustering() Clustering {
	return &kmeansClustering{
		km: kmeans.New(),
	}
}

var _ Clustering = new(kmeansClustering)

func (k *kmeansClustering) ComputeClusters(clusterCnt int64, data [][]float32) (centroids [][]float32, err error) {

	var input = make([]clusters.Observation, len(data))
	for r := 0; r < len(data); r++ {
		coord := make(clusters.Coordinates, len(data[r]))
		for c := 0; c < len(data[r]); c++ {
			coord[c] = float64(data[r][c])
		}
		input[r] = coord
	}

	partitions, err := k.km.Partition(input, int(clusterCnt))

	centroids = make([][]float32, len(partitions))
	for r, p := range partitions {
		center := make([]float32, len(p.Center))
		for c := 0; c < len(p.Center); c++ {
			center[c] = float32(p.Center[c])
		}
		centroids[r] = center
	}

	return centroids, nil
}

func (k *kmeansClustering) Close() {
	//TODO implement me
	panic("implement me")
}
