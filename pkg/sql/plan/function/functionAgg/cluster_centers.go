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

package functionAgg

import (
	"bytes"
	"encoding/json"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/agg"
	"github.com/matrixorigin/matrixone/pkg/sql/plan/function/functionAgg/algo/elkans_kmeans"
	"strconv"
	"strings"
)

const (
	defaultKmeansMaxIteration   = 500
	defaultKmeansDeltaThreshold = 0.01
)

var (
	AggClusterCentersSupportedParameters = []types.T{
		types.T_array_float32, types.T_array_float64,
	}

	AggClusterCentersReturnType = func(typs []types.Type) types.Type {
		return types.T_json.ToType()
	}
)

// NewAggClusterCenters this agg func will take list of vectors/arrays and run clustering algorithm like kmeans and
// return list of centroids for each cluster.
func NewAggClusterCenters(overloadID int64, dist bool, inputTypes []types.Type, outputType types.Type, config any, _ any) (agg.Agg[any], error) {
	aggPriv := &sAggClusterCenters{}

	var err error
	aggPriv.clusterCnt, aggPriv.distType, err = decodeConfig(config)
	if err != nil {
		return nil, err
	}

	switch inputTypes[0].Oid {
	case types.T_array_float32, types.T_array_float64:
		aggPriv.arrType = inputTypes[0]
		if dist {
			return agg.NewUnaryDistAgg(overloadID, aggPriv, false, inputTypes[0], outputType, aggPriv.Grows, aggPriv.Eval, aggPriv.Merge, aggPriv.Fill, nil), nil
		}
		return agg.NewUnaryAgg(overloadID, aggPriv, false, inputTypes[0], outputType, aggPriv.Grows, aggPriv.Eval, aggPriv.Merge, aggPriv.Fill, nil), nil
	}
	return nil, moerr.NewInternalErrorNoCtx("unsupported type '%s' for cluster_centers", inputTypes[0])
}

type sAggClusterCenters struct {
	// groupedData will hold the list of vectors/arrays. It is a 3D slice because it is a list of groups (based on group by)
	// and each group will have list of vectors/arrays
	// [group1] -> [array1, array2]
	// [group2] -> [array3, array4, array5]
	// NOTE: here array is []byte ie types.T_varchar
	groupedData [][][]byte

	clusterCnt int64
	distType   string

	// arrType is the type of the array/vector
	// It is used while converting array/vector to string and vice versa
	arrType types.Type
}

func (s *sAggClusterCenters) Grows(cnt int) {
	// grow the groupedData slice based on the number of groups
	s.groupedData = append(s.groupedData, make([][][]byte, cnt)...)
}
func (s *sAggClusterCenters) Free(_ *mpool.MPool) {}
func (s *sAggClusterCenters) Fill(groupNumber int64, values []byte, lastResult []byte, count int64, isEmpty bool, isNull bool) ([]byte, bool, error) {
	// NOTE: this function is copied from group_concat.go

	if isNull {
		return nil, isEmpty, nil
	}

	oneDimByteArrToTwoDimByteArr := func(data []byte, chunkSize int64) [][]byte {
		var chunks [][]byte

		for i := int64(0); i < int64(len(data)); i += chunkSize {
			end := i + chunkSize
			if end > int64(len(data)) {
				end = int64(len(data))
			}
			chunks = append(chunks, data[i:end])
		}
		return chunks
	}

	// values should be ideally having list of vectors/arrays
	arrays := oneDimByteArrToTwoDimByteArr(values, int64(len(values))/count)
	s.groupedData[groupNumber] = append(s.groupedData[groupNumber], arrays...)

	return nil, false, nil
}
func (s *sAggClusterCenters) Merge(groupNumber1 int64, groupNumber2 int64, result1 []byte, result2 []byte, isEmpty1 bool, isEmpty2 bool, priv2 any) ([]byte, bool, error) {
	// NOTE: this function is copied from group_concat.go

	if isEmpty2 {
		return nil, isEmpty1 && isEmpty2, nil
	}

	s2 := priv2.(*sAggClusterCenters)
	s.groupedData[groupNumber1] = append(s.groupedData[groupNumber1], s2.groupedData[groupNumber2][:]...)

	return nil, isEmpty1 && isEmpty2, nil
}

func (s *sAggClusterCenters) Eval(lastResult [][]byte) ([][]byte, error) {
	result := make([][]byte, 0, len(s.groupedData))

	// The kmeans logic
	for i := 0; i < len(s.groupedData); i++ {
		if len(s.groupedData[i]) == 0 {
			continue
		}

		if len(s.groupedData[i]) == 1 {
			jsonData, err := json.Marshal(s.groupedData[i])
			if err != nil {
				return nil, err
			}
			result = append(result, jsonData)
			continue
		}

		// each groupedData contains vectors based on group by clause.
		// if no group by is mentioned, we have groupedData[0] having all the data.
		var vecf64 [][]float64
		for _, arr := range s.groupedData[i] {
			switch s.arrType.Oid {
			case types.T_array_float32:
				// 1. convert []byte to []float64
				vecf32 := types.BytesToArray[float32](arr)

				// 1.a cast to []float64
				_vecf64 := make([]float64, len(vecf32))
				for j, v := range vecf32 {
					_vecf64[j] = float64(v)
				}

				vecf64 = append(vecf64, _vecf64)

			case types.T_array_float64:
				vecf64 = append(vecf64, types.BytesToArray[float64](arr))
			}
		}

		// 2. call kmeans.
		distanceType, err := s.findDistanceType()
		if err != nil {
			return nil, err
		}

		clusterer, err := elkans_kmeans.NewElkansKMeans(vecf64,
			int(s.clusterCnt),
			defaultKmeansMaxIteration,
			defaultKmeansDeltaThreshold,
			distanceType)
		if err != nil {
			return nil, err
		}

		centers, err := clusterer.Cluster()
		if err != nil {
			return nil, err
		}

		// 3. convert centroids to json string
		jsonData, err := json.Marshal(centers)
		if err != nil {
			return nil, err
		}

		// 4. convert json string to []byte
		result = append(result, jsonData)
	}

	return result, nil
}

func (s *sAggClusterCenters) findDistanceType() (elkans_kmeans.DistanceType, error) {
	var distanceType elkans_kmeans.DistanceType
	switch s.distType {
	case "L2", "":
		distanceType = elkans_kmeans.L2
	case "IP":
		distanceType = elkans_kmeans.InnerProduct
	case "COSINE":
		distanceType = elkans_kmeans.CosineDistance
	}
	return distanceType, moerr.NewInternalErrorNoCtx("unsupported distance function '%s' for cluster_centers", s.distType)
}

func (s *sAggClusterCenters) MarshalBinary() ([]byte, error) {

	if len(s.groupedData) == 0 {
		return nil, nil
	}

	var buf bytes.Buffer

	// arrType
	a, err := s.arrType.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(a)

	// groupedData
	strList := make([]string, 0, len(s.groupedData))
	for i := range s.groupedData {
		strList = append(strList, s.arraysToString(s.groupedData[i]))
	}
	buf.Write(types.EncodeStringSlice(strList))

	return buf.Bytes(), nil
}

func (s *sAggClusterCenters) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// arrType
	err := s.arrType.UnmarshalBinary(data[:20])
	if err != nil {
		return err
	}

	// groupedData
	data = data[20:]
	strList := types.DecodeStringSlice(data)
	s.groupedData = make([][][]byte, len(strList))
	for i := range s.groupedData {
		s.groupedData[i] = s.stringToArrays(strList[i])
	}
	return nil
}

// arraysToString converts list of array/vector to string
// e.g. []array -> "1,2,3|4,5,6|"
func (s *sAggClusterCenters) arraysToString(arrays [][]byte) string {
	var res string
	var commaSeperatedArrString string
	for _, arr := range arrays {
		switch s.arrType.Oid {
		case types.T_array_float32:
			commaSeperatedArrString = types.BytesToArrayToString[float32](arr)
		case types.T_array_float64:
			commaSeperatedArrString = types.BytesToArrayToString[float64](arr)
		}
		res += commaSeperatedArrString + "|"
	}
	return res

}

// stringToArrays converts string to a list of array/vector
// e.g. "1,2,3|4,5,6|" -> []array
func (s *sAggClusterCenters) stringToArrays(str string) [][]byte {
	arrays := strings.Split(str, "|")
	var res [][]byte
	var array []byte
	for _, arr := range arrays {
		if len(strings.TrimSpace(arr)) == 0 {
			continue
		}
		switch s.arrType.Oid {
		case types.T_array_float32:
			array, _ = types.StringToArrayToBytes[float32](arr)
		case types.T_array_float64:
			array, _ = types.StringToArrayToBytes[float64](arr)
		}
		res = append(res, array)
	}
	return res

}

func decodeConfig(config any) (k int64, distFn string, err error) {
	bts, ok := config.([]byte)
	if ok && bts != nil {
		commaSeperatedConfigStr := string(bts)
		configs := strings.Split(commaSeperatedConfigStr, ",")
		if len(configs) == 1 {
			k, err = strconv.ParseInt(strings.TrimSpace(configs[0]), 10, 64)
			if err != nil {
				return 0, "", err
			}
			return k, "L2", nil
		}

		if len(configs) == 2 {
			k, err = strconv.ParseInt(strings.TrimSpace(configs[0]), 10, 64)
			if err != nil {
				return 0, "", err
			}

			distFn = strings.TrimSpace(configs[1])
			return k, distFn, nil
		}

	}
	return 1, "L2", nil
}
