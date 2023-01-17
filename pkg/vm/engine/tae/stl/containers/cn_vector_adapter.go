package containers

import (
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	cnVector "github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/stl"
	"io"
	"unsafe"
)

type CnStlVector[T any] struct {
	downstreamVector *cnVector.Vector
	mpool            *mpool.MPool
}

func NewStlVector[T any](opts ...Options) *CnStlVector[T] {
	vec := &CnStlVector[T]{
		//downstreamVector: cnVector.New(types.DecodeType()),
		mpool: opts[0].Allocator,
	}

	return vec
}

func (vec CnStlVector[T]) Append(v T) {
	err := vec.downstreamVector.Append(v, false, vec.mpool)
	if err != nil {
		return
	}
}

func (vec CnStlVector[T]) Data() []byte {
	data, _ := vec.downstreamVector.Show()
	return data
}

func (vec CnStlVector[T]) Length() int {
	return vec.downstreamVector.Length()
}

func (vec CnStlVector[T]) Get(i int) (v T) {
	//return vec.downstreamVector.GetBytes(i)
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) String() string {
	return vec.downstreamVector.String()
}

func (vec CnStlVector[T]) Allocated() int {
	return vec.downstreamVector.Length()
}

func (vec CnStlVector[T]) Close() {
	vec.downstreamVector.Free(vec.mpool)
}

func (vec CnStlVector[T]) Clone(offset, length int, allocator ...*mpool.MPool) stl.Vector[T] {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) ReadBytes(data *stl.Bytes, share bool) {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) Reset() {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) IsView() bool {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) Bytes() *stl.Bytes {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) WindowAsBytes(offset, length int) *stl.Bytes {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) DataWindow(offset, length int) []byte {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) Slice() []T {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) SlicePtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) SliceWindow(offset, length int) []T {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) AppendMany(vals ...T) {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) Update(i int, v T) {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) Delete(i int) (deleted T) {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) BatchDelete(rowGen common.RowGen, cnt int) {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) BatchDeleteInts(sels ...int) {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) BatchDeleteUint32s(sels ...uint32) {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) GetAllocator() *mpool.MPool {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) Capacity() int {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) Desc() string {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) WriteTo(writer io.Writer) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) ReadFrom(reader io.Reader) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (vec CnStlVector[T]) InitFromSharedBuf(buf []byte) (int64, error) {
	//TODO implement me
	panic("implement me")
}
