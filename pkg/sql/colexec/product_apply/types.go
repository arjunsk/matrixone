// Copyright 2021 Matrix Origin
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

package productapply

import (
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/common/reuse"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

var _ vm.Operator = new(ProductApply)

const (
	Build = iota
	Probe
	End
)

type container struct {
	colexec.ReceiverOperator

	state    int
	probeIdx int
	bat      *batch.Batch
	rbat     *batch.Batch
	inBat    *batch.Batch
}

type ProductApply struct {
	ctr       *container
	Typs      []types.Type
	Result    []colexec.ResultPos
	IsShuffle bool
	vm.OperatorBase
}

func (productApply *ProductApply) GetOperatorBase() *vm.OperatorBase {
	return &productApply.OperatorBase
}

func init() {
	reuse.CreatePool[ProductApply](
		func() *ProductApply {
			return &ProductApply{}
		},
		func(a *ProductApply) {
			*a = ProductApply{}
		},
		reuse.DefaultOptions[ProductApply]().
			WithEnableChecker(),
	)
}

func (productApply ProductApply) TypeName() string {
	return opName
}

func NewArgument() *ProductApply {
	return reuse.Alloc[ProductApply](nil)
}

func (productApply *ProductApply) Release() {
	if productApply != nil {
		reuse.Free[ProductApply](productApply, nil)
	}
}

func (productApply *ProductApply) Reset(proc *process.Process, pipelineFailed bool, err error) {
	productApply.Free(proc, pipelineFailed, err)
}

func (productApply *ProductApply) Free(proc *process.Process, pipelineFailed bool, err error) {
	ctr := productApply.ctr
	if ctr != nil {
		mp := proc.Mp()
		ctr.cleanBatch(mp)
		ctr.FreeAllReg()
		productApply.ctr = nil
	}
}

func (ctr *container) cleanBatch(mp *mpool.MPool) {
	if ctr.bat != nil {
		ctr.bat.Clean(mp)
		ctr.bat = nil
	}
	if ctr.rbat != nil {
		ctr.rbat.Clean(mp)
		ctr.rbat = nil
	}
	if ctr.inBat != nil {
		ctr.inBat.Clean(mp)
		ctr.inBat = nil
	}
}
