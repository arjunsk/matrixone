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

package containers

import (
	"bytes"
	"github.com/RoaringBitmap/roaring"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	cnNulls "github.com/matrixorigin/matrixone/pkg/container/nulls"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/stl"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/stl/containers"
	"io"
)

type Options = containers.Options
type Bytes = stl.Bytes

// var DefaultAllocator = alloc.NewAllocator(int(common.G) * 100)

type ItOp = func(v any, row int) error

type VectorView interface {
	Nullable() bool
	IsNull(i int) bool
	HasNull() bool
	NullMask() *cnNulls.Nulls

	Bytes() *Bytes
	Slice() any
	Get(i int) any

	Length() int
	Capacity() int
	Allocated() int
	GetAllocator() *mpool.MPool
	GetType() types.Type
	String() string
	PPString(num int) string

	Foreach(op ItOp, sels *roaring.Bitmap) error
	ForeachWindow(offset, length int, op ItOp, sels *roaring.Bitmap) error

	WriteTo(w io.Writer) (int64, error)
}

type Vector interface {
	VectorView
	ResetWithData(bs *Bytes, nulls *cnNulls.Nulls)
	Update(i int, v any)
	Delete(i int)
	Compact(*roaring.Bitmap)
	Append(v any)
	AppendMany(vs ...any)
	Extend(o Vector)
	ExtendWithOffset(src Vector, srcOff, srcLen int)
	CloneWindow(offset, length int, allocator ...*mpool.MPool) Vector

	Equals(o Vector) bool
	Window(offset, length int) Vector
	WriteTo(w io.Writer) (int64, error)
	ReadFrom(r io.Reader) (int64, error)

	ReadFromFile(common.IVFile, *bytes.Buffer) error

	Close()
}

type Batch struct {
	Attrs   []string
	Vecs    []Vector
	Deletes *roaring.Bitmap
	nameidx map[string]int
	// refidx  map[int]int
}
