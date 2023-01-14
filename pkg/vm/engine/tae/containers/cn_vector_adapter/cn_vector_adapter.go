// Copyright 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cn_vector_adapter

import (
	"bytes"
	"github.com/RoaringBitmap/roaring"
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	cnVector "github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/containers"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/stl"
	"io"
	"unsafe"
)

type CnVector[T any] struct {
	downstreamVector *cnVector.Vector
}

//var _ stl.CnVector[T] = new(CnVector[T])
//var _ containers.CnVector = new(CnVector[T])

func NewContainerVector[T any](typ types.Type, nullable bool, opts ...containers.Options) *CnVector[T] {
	vec := &CnVector[T]{
		downstreamVector: cnVector.New(typ),
	}
	return vec
}

func NewStlVector[T any](opts ...containers.Options) *CnVector[T] {
	vec := &CnVector[T]{}
	return vec
}

// ************* TAE Container CnVector *************

func (vec CnVector[T]) IsView() bool {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Nullable() bool {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) IsNull(i int) bool {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) HasNull() bool {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) NullMask() *roaring64.Bitmap {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Data() []byte {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Bytes() *containers.Bytes {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Slice() any {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) SlicePtr() unsafe.Pointer {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) DataWindow(offset, length int) []byte {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Get(i int) any {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Length() int {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Capacity() int {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Allocated() int {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) GetAllocator() *mpool.MPool {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) GetType() types.Type {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) String() string {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) PPString(num int) string {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Foreach(op containers.ItOp, sels *roaring.Bitmap) error {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) ForeachWindow(offset, length int, op containers.ItOp, sels *roaring.Bitmap) error {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) WriteTo(w io.Writer) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Reset() {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) ResetWithData(bs *containers.Bytes, nulls *roaring64.Bitmap) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) GetView() containers.VectorView {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Update(i int, v any) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Delete(i int) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Compact(bitmap *roaring.Bitmap) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Append(v any) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) AppendMany(vs ...any) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) AppendNoNulls(s any) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Extend(o containers.Vector) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) ExtendWithOffset(src containers.Vector, srcOff, srcLen int) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) CloneWindow(offset, length int, allocator ...*mpool.MPool) containers.Vector {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Equals(o containers.Vector) bool {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Window(offset, length int) containers.Vector {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) ReadFrom(r io.Reader) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) ReadFromFile(file common.IVFile, buffer *bytes.Buffer) error {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Close() {
	//TODO implement me
	panic("implement me")
}

// ************* STL Container CnVector *************
// Will remove Stl suffix, deprecate the aliases.

func (vec CnVector[T]) CloneStl(offset, length int, allocator ...*mpool.MPool) stl.Vector[T] {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) ReadBytesStl(data *stl.Bytes, share bool) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) WindowAsBytes(offset, length int) *stl.Bytes {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) SliceStl() []T {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) SliceWindowStl(offset, length int) []T {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) GetStl(i int) (v T) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) AppendStl(v T) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) AppendManyStl(vals ...T) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) UpdateStl(i int, v T) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) DeleteStl(i int) (deleted T) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) BatchDelete(rowGen common.RowGen, cnt int) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) BatchDeleteInts(sels ...int) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) BatchDeleteUint32s(sels ...uint32) {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) Desc() string {
	//TODO implement me
	panic("implement me")
}

func (vec CnVector[T]) InitFromSharedBuf(buf []byte) (int64, error) {
	//TODO implement me
	panic("implement me")
}
