// Copyright 2022 Matrix Origin
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

package preinsertsecondaryindex

import (
	"bytes"

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/nulls"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/sql/util"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

const (
	indexColPos int32 = iota
	pkColPos
	skPkColPos
	rowIdColPos
)

func String(_ any, buf *bytes.Buffer) {
	buf.WriteString("pre processing insert secondary key")
}

func Prepare(_ *process.Process, _ any) error {
	return nil
}

func Call(idx int, proc *process.Process, arg any, _, _ bool) (process.ExecStatus, error) {
	argument := arg.(*Argument)

	analy := proc.GetAnalyze(idx)
	analy.Start()
	defer analy.Stop()

	inputBat := proc.InputBatch()
	if inputBat == nil {
		return process.ExecStop, nil
	}

	if inputBat.IsEmpty() {
		proc.PutBatch(inputBat)
		proc.SetInputBatch(batch.EmptyBatch)
		return process.ExecNext, nil
	}
	defer proc.PutBatch(inputBat)

	var vec *vector.Vector
	var bitMap *nulls.Nulls

	secondaryColumnPos := argument.PreInsertCtx.Columns
	pkPos := int(argument.PreInsertCtx.PkColumn)
	// tableDef := argument.PreInsertCtx.TableDef

	var insertSecondaryBat *batch.Batch
	isUpdate := inputBat.Vecs[len(inputBat.Vecs)-1].GetType().Oid == types.T_Rowid
	if isUpdate {
		insertSecondaryBat = batch.NewWithSize(3 + 1)
		insertSecondaryBat.Attrs = []string{catalog.IndexTableIndexColName, catalog.IndexTablePrimaryColName, catalog.IndexTableSkPrimaryColName, catalog.Row_ID}
	} else {
		insertSecondaryBat = batch.NewWithSize(2 + 1)
		insertSecondaryBat.Attrs = []string{catalog.IndexTableIndexColName, catalog.IndexTablePrimaryColName, catalog.IndexTableSkPrimaryColName}
	}

	colCount := len(secondaryColumnPos)

	if colCount == 1 {
		pos := secondaryColumnPos[indexColPos]
		vec, bitMap = util.CompactSingleIndexCol(inputBat.Vecs[pos], proc)
	} else {
		vs := make([]*vector.Vector, colCount)
		for vIdx, pIdx := range secondaryColumnPos {
			vs[vIdx] = inputBat.Vecs[pIdx]
		}
		vec, bitMap = util.SerialWithCompacted(vs, proc)
	}
	insertSecondaryBat.SetVector(indexColPos, vec)
	insertSecondaryBat.SetRowCount(vec.Length())

	vecPk := util.CompactPrimaryCol(inputBat.Vecs[pkPos], bitMap, proc)
	insertSecondaryBat.SetVector(pkColPos, vecPk)

	{ // 3rd column ie sk's + original_pk
		vs := make([]*vector.Vector, colCount+1)
		for vIdx, pIdx := range secondaryColumnPos {
			vs[vIdx] = inputBat.Vecs[pIdx]
		}
		vs[colCount] = vecPk
		vec, _ = util.SerialWithCompacted(vs, proc)
		insertSecondaryBat.SetVector(skPkColPos, vec)
	}

	if isUpdate {
		rowIdInBat := len(inputBat.Vecs) - 1
		insertSecondaryBat.SetVector(rowIdColPos, proc.GetVector(*inputBat.GetVector(int32(rowIdInBat)).GetType()))
		err := insertSecondaryBat.Vecs[rowIdColPos].UnionBatch(inputBat.Vecs[rowIdInBat], 0, inputBat.Vecs[rowIdInBat].Length(), nil, proc.Mp())
		if err != nil {
			insertSecondaryBat.Clean(proc.GetMPool())
			return process.ExecNext, err
		}
	}
	proc.SetInputBatch(insertSecondaryBat)
	return process.ExecNext, nil
}
