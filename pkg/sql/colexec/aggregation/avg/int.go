package avg

import (
	"matrixbase/pkg/container/types"
	"matrixbase/pkg/container/vector"
	"matrixbase/pkg/encoding"
	"matrixbase/pkg/sql/colexec/aggregation"
	"matrixbase/pkg/vectorize/sum"
	"matrixbase/pkg/vm/mempool"
	"matrixbase/pkg/vm/process"
)

func NewInt(typ types.Type) *intAvg {
	return &intAvg{typ: typ}
}

func (a *intAvg) Reset() {
	a.cnt = 0
	a.sum = 0
}

func (a *intAvg) Type() types.Type {
	return a.typ
}

func (a *intAvg) Dup() aggregation.Aggregation {
	return &intAvg{typ: a.typ}
}

func (a *intAvg) Fill(sels []int64, vec *vector.Vector) error {
	if n := len(sels); n > 0 {
		switch vec.Typ.Oid {
		case types.T_int8:
			a.sum += sum.Int8SumSels(vec.Col.([]int8), sels)
		case types.T_int16:
			a.sum += sum.Int16SumSels(vec.Col.([]int16), sels)
		case types.T_int32:
			a.sum += sum.Int32SumSels(vec.Col.([]int32), sels)
		case types.T_int64:
			a.sum += sum.Int64SumSels(vec.Col.([]int64), sels)
		}
		a.cnt += int64(n - vec.Nsp.FilterCount(sels))
	} else {
		switch vec.Typ.Oid {
		case types.T_int8:
			a.sum += sum.Int8Sum(vec.Col.([]int8))
		case types.T_int16:
			a.sum += sum.Int16Sum(vec.Col.([]int16))
		case types.T_int32:
			a.sum += sum.Int32Sum(vec.Col.([]int32))
		case types.T_int64:
			a.sum += sum.Int64Sum(vec.Col.([]int64))
		}
		a.cnt += int64(vec.Length() - vec.Nsp.Length())
	}
	return nil
}

func (a *intAvg) Eval() interface{} {
	if a.cnt == 0 {
		return nil
	}
	return float64(a.sum) / float64(a.cnt)
}

func (a *intAvg) EvalCopy(proc *process.Process) (*vector.Vector, error) {
	data, err := proc.Alloc(8)
	if err != nil {
		return nil, err
	}
	vec := vector.New(a.typ)
	if a.cnt == 0 {
		vec.Nsp.Add(0)
		copy(data[mempool.CountSize:], encoding.EncodeFloat64(0))
	} else {
		copy(data[mempool.CountSize:], encoding.EncodeFloat64(float64(a.sum)/float64(a.cnt)))
	}
	vec.Data = data
	vec.Col = encoding.DecodeFloat64Slice(data[mempool.CountSize : mempool.CountSize+8])
	return vec, nil
}
