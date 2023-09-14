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

/*
#cgo LDFLAGS: -lfaiss_c

#include <stdlib.h>
#include <faiss/c_api/Clustering_c.h>
#include <faiss/c_api/impl/AuxIndexStructures_c.h>
#include <faiss/c_api/index_factory_c.h>
#include <faiss/c_api/error_c.h>
*/
import "C"
import (
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
)

// CGO code for https://github.com/facebookresearch/faiss/blob/main/c_api/Clustering_c.h functions
// K-means : https://github.com/facebookresearch/faiss/wiki/Faiss-building-blocks:-clustering,-PCA,-quantization
type faissClustering struct {
}

var _ Clustering = new(faissClustering)

func NewFaissClustering() Clustering {
	return &faissClustering{}
}

func (f *faissClustering) ComputeClusters(clusterCnt int64, data [][]float32) (centroids [][]float32, err error) {
	if len(data) == 0 {
		return nil, moerr.NewInternalErrorNoCtx("empty rows")
	}
	if len(data[0]) == 0 {
		return nil, moerr.NewInternalErrorNoCtx("zero dimensions")
	}

	rowCnt := int64(len(data))
	dims := int64(len(data[0]))

	// flatten data from 2D to 1D
	vectorFlat := make([]float32, dims*rowCnt)

	//TODO: optimize
	for r := int64(0); r < rowCnt; r++ {
		for c := int64(0); c < dims; c++ {
			vectorFlat[(r*dims)+c] = data[r][c]
		}
	}

	//TODO: do memory de-allocation if any
	centroidsFlat := make([]float32, dims*clusterCnt)
	var qError float32
	c := C.faiss_kmeans_clustering(
		C.ulong(dims),                 // d dimension of the data
		C.ulong(rowCnt),               // n nb of training vectors
		C.ulong(clusterCnt),           // k nb of output centroids
		(*C.float)(&vectorFlat[0]),    // x training set (size n * d)
		(*C.float)(&centroidsFlat[0]), // centroids output centroids (size k * d)
		(*C.float)(&qError),           // q_error final quantization error
		//@return error code
	)
	if c != 0 {
		return nil, getLastError()
	}

	if qError <= 0 {
		//final quantization error
		return nil, moerr.NewInternalErrorNoCtx("final quantization error >0")
	}

	centroids = make([][]float32, clusterCnt)
	for r := int64(0); r < clusterCnt; r++ {
		centroids[r] = centroidsFlat[r*dims : (r+1)*dims]
	}
	return
}

func (f *faissClustering) Close() {
	//TODO implement me
	panic("implement me")
}

func getLastError() error {
	return moerr.NewInternalErrorNoCtx(C.GoString(C.faiss_get_last_error()))
}
