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

package elkans

import (
	"fmt"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/sql/plan/function/functionAgg/algos/kmeans"
	"golang.org/x/sync/errgroup"
	"math"
	"math/rand"
	"sync"
)

// ElkanClusterer is an improved kmeans algorithm which using the triangle inequality to reduce the number of
// distance calculations. As quoted from the paper:
// "The main contribution of this paper is an optimized version of the standard k-means method, with which the number
// of distance computations is in practice closer to `n` than to `nke`, where n is the number of vectors, k is the number
// of centroids, and e is the number of iterations needed until convergence."
//
// However, during each iteration of the algorithm, the lower bounds l(x, c) are updated for all points x and centers c.
// These updates take O(nk) time, so the complexity of the algorithm remains at least O(nke), even though the number
// of distance calculations is roughly O(n) only.
// NOTE that, distance calculation is very expensive for higher dimension vectors.
//
// Ref Paper: https://cdn.aaai.org/ICML/2003/ICML03-022.pdf
// Ref Material :https://www.cse.iitd.ac.in/~rjaiswal/2015/col870/Project/Nipun.pdf
type ElkanClusterer struct {

	// for each of the n vectors, we keep track of the following data
	vectorList  [][]float64
	vectorMetas []vectorMeta
	assignments []int

	// for each of the k centroids, we keep track of the following data
	Centroids                   [][]float64
	halfInterCentroidDistMatrix [][]float64
	minHalfInterCentroidDist    []float64

	// thresholds
	maxIterations  int     // e in paper
	deltaThreshold float64 // used for early convergence

	// counts
	clusterCnt int // k in paper
	vectorCnt  int // n in paper

	distFn   kmeans.DistanceFunction
	initType kmeans.InitType
	rand     *rand.Rand
}

// vectorMeta holds info required for Elkan's kmeans optimization.
// lower is the lower bound distance of a vector to `each` of its centroids. Hence, there are k values.
// upper is the upper bound distance of a vector to its `closest` centroid. Hence, only one value.
// recompute is a flag to indicate if the distance to centroids needs to be recomputed. if false, then use upper.
type vectorMeta struct {
	lower     []float64
	upper     float64
	recompute bool
}

var _ kmeans.Clusterer = new(ElkanClusterer)

func NewKMeans(vectors [][]float64,
	clusterCnt, maxIterations int,
	deltaThreshold float64,
	distanceType kmeans.DistanceType,
	initType kmeans.InitType,
) (kmeans.Clusterer, error) {

	err := validateArgs(vectors, clusterCnt, maxIterations, deltaThreshold, distanceType, initType)
	if err != nil {
		return nil, err
	}

	assignments := make([]int, len(vectors))
	var metas = make([]vectorMeta, len(vectors))
	for i := range metas {
		metas[i] = vectorMeta{
			lower:     make([]float64, clusterCnt),
			upper:     0,
			recompute: true,
		}
	}

	centroidDist := make([][]float64, clusterCnt)
	minCentroidDist := make([]float64, clusterCnt)
	for i := range centroidDist {
		centroidDist[i] = make([]float64, clusterCnt)
	}

	distanceFunction, err := resolveDistanceFn(distanceType)
	if err != nil {
		return nil, err
	}

	return &ElkanClusterer{
		maxIterations:  maxIterations,
		deltaThreshold: deltaThreshold,

		vectorList:  vectors,
		assignments: assignments,
		vectorMetas: metas,

		//Centroids will be initialized by InitCentroids()
		halfInterCentroidDistMatrix: centroidDist,
		minHalfInterCentroidDist:    minCentroidDist,

		distFn:     distanceFunction,
		initType:   initType,
		clusterCnt: clusterCnt,
		vectorCnt:  len(vectors),

		rand: rand.New(rand.NewSource(kmeans.DefaultRandSeed)),
	}, nil
}

// InitCentroids initializes the centroids using initialization algorithms like kmeans++ or random.
func (km *ElkanClusterer) InitCentroids() {
	var initializer Initializer
	switch km.initType {
	case kmeans.KmeansPlusPlus:
		initializer = NewKMeansPlusPlusInitializer(km.distFn)
	case kmeans.Random:
		initializer = NewRandomInitializer()
	default:
		initializer = NewRandomInitializer()
	}
	km.Centroids = initializer.InitCentroids(km.vectorList, km.clusterCnt)
}

// Cluster returns the final centroids and the error if any.
func (km *ElkanClusterer) Cluster() ([][]float64, error) {

	if km.vectorCnt == km.clusterCnt {
		return km.vectorList, nil
	}

	km.InitCentroids() // step 0.a
	km.initBounds()    // step 0.b

	return km.elkansCluster()
}

func (km *ElkanClusterer) elkansCluster() ([][]float64, error) {

	for iter := 0; ; iter++ {
		km.computeCentroidDistances() // step 1

		changes := km.assignData() // step 2 and 3

		newCentroids := km.recalculateCentroids() // step 4

		km.updateBounds(newCentroids) // step 5 and 6

		km.Centroids = newCentroids // step 7

		if iter != 0 && km.isConverged(iter, changes) {
			break
		}
	}
	return km.Centroids, nil
}

func validateArgs(vectorList [][]float64, clusterCnt,
	maxIterations int, deltaThreshold float64,
	distanceType kmeans.DistanceType, initType kmeans.InitType) error {
	if vectorList == nil || len(vectorList) == 0 || len(vectorList[0]) == 0 {
		return moerr.NewInternalErrorNoCtx("input vectors is empty")
	}
	if clusterCnt > len(vectorList) {
		return moerr.NewInternalErrorNoCtx("cluster count is larger than vector count")
	}
	if maxIterations < 0 {
		return moerr.NewInternalErrorNoCtx("max iteration is out of bounds (must be >= 0)")
	}
	if deltaThreshold <= 0.0 || deltaThreshold >= 1.0 {
		return moerr.NewInternalErrorNoCtx("delta threshold is out of bounds (must be > 0.0 and < 1.0)")
	}
	if distanceType < 0 || distanceType > 2 {
		return moerr.NewInternalErrorNoCtx("distance type is not supported")
	}
	if initType < 0 || initType > 1 {
		return moerr.NewInternalErrorNoCtx("init type is not supported")
	}

	// validate that all vectors have the same dimension
	vecDim := len(vectorList[0])
	eg := new(errgroup.Group)
	for i := 1; i < len(vectorList); i++ {
		func(idx int) {
			eg.Go(func() error {
				if len(vectorList[idx]) != vecDim {
					return moerr.NewInternalErrorNoCtx(fmt.Sprintf("dim mismatch. "+
						"vector[%d] has dimension %d, "+
						"but vector[0] has dimension %d", idx, len(vectorList[idx]), vecDim))
				}
				return nil
			})
		}(i)
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

// initBounds initializes the lower bounds, upper bound and assignment for each vector.
func (km *ElkanClusterer) initBounds() {
	for x := range km.vectorList {
		minDist := math.MaxFloat64
		closestCenter := 0
		for c := 0; c < len(km.Centroids); c++ {
			dist := km.distFn(km.vectorList[x], km.Centroids[c])
			km.vectorMetas[x].lower[c] = dist
			if dist < minDist {
				minDist = dist
				closestCenter = c
			}
		}

		km.vectorMetas[x].upper = minDist
		km.assignments[x] = closestCenter
	}
}

// computeCentroidDistances computes the centroid distances and the min centroid distances.
// NOTE: here we are save 0.5 of centroid distance to avoid 0.5 multiplication in step 3(iii) and 3.b.
func (km *ElkanClusterer) computeCentroidDistances() {

	// step 1.a
	// For all centers c and c', compute 0.5 x d(c, c').
	var wg sync.WaitGroup
	for r := 0; r < km.clusterCnt; r++ {
		for c := r + 1; c < km.clusterCnt; c++ {
			wg.Add(1)
			go func(i, j int) {
				defer wg.Done()
				dist := 0.5 * km.distFn(km.Centroids[i], km.Centroids[j])
				km.halfInterCentroidDistMatrix[i][j] = dist
				km.halfInterCentroidDistMatrix[j][i] = dist
			}(r, c)
		}
	}
	wg.Wait()

	// step 1.b
	//  For all centers c, compute s(c)=0.5 x min{d(c, c') | c'!= c}.
	for i := 0; i < km.clusterCnt; i++ {
		currMinDist := math.MaxFloat64
		for j := 0; j < km.clusterCnt; j++ {
			if i == j {
				continue
			}
			currMinDist = math.Min(currMinDist, km.halfInterCentroidDistMatrix[i][j])
		}
		km.minHalfInterCentroidDist[i] = currMinDist
	}
}

// assignData assigns each vector to the nearest centroid.
// This is the place where most of the pruning happens.
func (km *ElkanClusterer) assignData() int {

	var ux float64 // currVecUpperBound
	var cx int     // currVecClusterAssignmentIdx
	changes := 0

	for x := range km.vectorList { // x is currVectorIdx

		ux = km.vectorMetas[x].upper // u(x) in the paper
		cx = km.assignments[x]       // c(x) in the paper

		// step 2 u(x) <= s(c(x))
		if ux <= km.minHalfInterCentroidDist[cx] {
			continue
		}

		for c := range km.Centroids { // c is nextPossibleCentroidIdx
			// step 3
			// For all remaining points x and centers c such that
			// (i) c != c(x) and
			// (ii) u(x)>l(x, c) and
			// (iii) u(x)> 0.5 x d(c(x), c)
			// NOTE: we proactively use the otherwise case
			if c == cx || // (i)
				ux <= km.vectorMetas[x].lower[c] || // ii)
				ux <= km.halfInterCentroidDistMatrix[cx][c] { // (iii)
				continue
			}

			//step 3.a - Bounds update
			// If r(x) then compute d(x, c(x)) and assign r(x)= false. Otherwise, d(x, c(x))=u(x).
			dxcx := ux // d(x, c(x)) in the paper ie distToCentroid
			if km.vectorMetas[x].recompute {
				dxcx = km.distFn(km.vectorList[x], km.Centroids[cx])
				km.vectorMetas[x].upper = dxcx
				km.vectorMetas[x].lower[cx] = dxcx
				km.vectorMetas[x].recompute = false
			}

			//step 3.b - Update
			// If d(x, c(x))>l(x, c) or d(x, c(x))> 0.5 d(c(x), c) then
			// Compute d(x, c)
			// If d(x, c)<d(x, c(x)) then assign c(x)=c.
			if dxcx > km.vectorMetas[x].lower[c] ||
				dxcx > km.halfInterCentroidDistMatrix[cx][c] {

				dxc := km.distFn(km.vectorList[x], km.Centroids[c]) // d(x,c) in the paper
				km.vectorMetas[x].lower[c] = dxc
				if dxc < dxcx {
					km.vectorMetas[x].upper = dxc

					cx = c
					km.assignments[x] = c
					changes++
				}
			}
		}
	}
	return changes
}

// recalculateCentroids calculates the new mean centroids based on the new assignments.
func (km *ElkanClusterer) recalculateCentroids() [][]float64 {
	clusterMembersCount := make([]int64, len(km.Centroids))
	clusterMembersDimWiseSum := make([][]float64, len(km.Centroids))

	for c := range km.Centroids {
		clusterMembersDimWiseSum[c] = make([]float64, len(km.vectorList[0]))
	}

	for x, vec := range km.vectorList {
		cx := km.assignments[x]
		clusterMembersCount[cx]++
		for dim := range vec {
			clusterMembersDimWiseSum[cx][dim] += vec[dim]
		}
	}

	newCentroids := append([][]float64{}, km.Centroids...)
	for c, newCentroid := range newCentroids {
		memberCnt := float64(clusterMembersCount[c])

		if memberCnt == 0 {
			// if the cluster is empty, reinitialize it to a random vector, since you can't find the mean of an empty set
			for l := range newCentroid {
				newCentroid[l] = 10 * (km.rand.Float64() - 0.5)
			}
		} else {
			// find the mean of the cluster members
			for dim := range newCentroid {
				newCentroid[dim] = clusterMembersDimWiseSum[c][dim] / memberCnt
			}
		}

	}

	return newCentroids
}

// updateBounds updates the lower and upper bounds for each vector.
func (km *ElkanClusterer) updateBounds(newCentroid [][]float64) {

	// compute the centroid shift distance matrix once.
	centroidShiftDist := make([]float64, km.clusterCnt)
	var wg sync.WaitGroup
	for c := 0; c < km.clusterCnt; c++ {
		wg.Add(1)
		go func(cIdx int) {
			defer wg.Done()
			centroidShiftDist[cIdx] = km.distFn(km.Centroids[cIdx], newCentroid[cIdx])
		}(c)
	}
	wg.Wait()

	// step 5
	//For each point x and center c, assign
	// l(x, c)= max{ l(x, c)-d(c, m(c)), 0 }
	for x := range km.vectorList {
		for c := range km.Centroids {
			shift := km.vectorMetas[x].lower[c] - centroidShiftDist[c]
			km.vectorMetas[x].lower[c] = math.Max(shift, 0)
		}

		// step 6
		// For each point x, assign
		// u(x)=u(x)+d(m(c(x)), c(x))
		// r(x)= true
		cx := km.assignments[x] // ie currVecClusterAssignmentIdx
		km.vectorMetas[x].upper += centroidShiftDist[cx]
		km.vectorMetas[x].recompute = true
	}
}

// isConverged checks if the algorithm has converged.
func (km *ElkanClusterer) isConverged(iter int, changes int) bool {
	if iter == km.maxIterations ||
		changes < int(float64(km.vectorCnt)*km.deltaThreshold) ||
		changes == 0 {
		return true
	}
	return false
}
