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

package plan

import (
	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/sql/parsers/tree"
)

func buildAlterTableReindex(stmt *tree.AlterTable, ctx CompilerContext) (*Plan, error) {
	if len(stmt.Options) != 1 {
		return nil, moerr.NewInternalErrorNoCtx("currently we only support reindexing one column at a time")
	}

	// 1. table + database name
	tableName := string(stmt.Table.ObjectName)
	databaseName := string(stmt.Table.SchemaName)
	if databaseName == "" {
		databaseName = ctx.DefaultDatabase()
	}

	// 2. tableDef
	_, tableDef := ctx.Resolve(databaseName, tableName)
	if tableDef == nil {
		return nil, moerr.NewNoSuchTable(ctx.GetContext(), databaseName, tableName)
	}

	oriPriKeyColName := getTablePriKeyName(tableDef.Pkey)
	if oriPriKeyColName == "" {
		return nil, moerr.NewInternalErrorNoCtx("primary key cannot be empty")
	}

	// 3. alterTable init
	alterTableReIndex := &plan.AlterTable{
		AlgorithmType: plan.AlterTable_REINDEX,
	}

	// 4. reindex params
	reIndexDef := stmt.Options[0].(*tree.AlterOptionReindex)
	_, _, reIndexColName := reIndexDef.ColumnName.GetNames()
	_, reIndexColType := getSecKeyPos(tableDef, reIndexColName)
	reIndexAlgo := reIndexDef.KeyType

	switch reIndexAlgo {
	case tree.INDEX_TYPE_IVFFLAT:
		alterTableReIndex.Actions = make([]*plan.AlterTable_Action, 2)
		{
			// Action 1 - rebuild

			// 1. find the indexTableName
			found := false
			var indexTableName string
			for _, indexdef := range tableDef.Indexes {
				if len(indexdef.Parts[0]) == 1 &&
					indexdef.Parts[0] == reIndexColName &&
					indexdef.IndexAlgo == tree.INDEX_TYPE_IVFFLAT.ToString() &&
					indexdef.IndexAlgoTableType == catalog.SystemSecondaryIndex_IvfCentroidsRel {
					indexTableName = indexdef.IndexTableName
					found = true
					break
				}
			}
			if !found {
				return nil, moerr.NewInternalErrorNoCtx("aux1 table not found")
			}

			alterTableReIndex.Actions[0] = &plan.AlterTable_Action{
				Action: &plan.AlterTable_Action_ReindexBuildCol{
					ReindexBuildCol: &plan.AlterReindexBuildCol{
						DbName:                  databaseName,
						TableName:               tableName,
						IndexTableName:          indexTableName,
						OriginTableIndexColName: reIndexColName,
						OriginTableIndexColType: reIndexColType,
						IndexAlgo:               reIndexAlgo.ToString(),
					},
				},
			}
		}

		{
			// Action 2 - remap

			// 0. find the indexTableName
			found := false
			var sourceIndexTableName string
			for _, indexdef := range tableDef.Indexes {
				if len(indexdef.Parts[0]) == 1 &&
					indexdef.Parts[0] == reIndexColName &&
					indexdef.IndexAlgo == tree.INDEX_TYPE_IVFFLAT.ToString() &&
					indexdef.IndexAlgoTableType == catalog.SystemSecondaryIndex_IvfCentroidsRel {
					sourceIndexTableName = indexdef.IndexTableName
					found = true
					break
				}
			}
			if !found {
				return nil, moerr.NewInternalErrorNoCtx("aux2 table not found")
			}

			// 1. find the indexTableName
			found = false
			var destIndexTableName string
			for _, indexdef := range tableDef.Indexes {
				if len(indexdef.Parts[0]) == 1 &&
					indexdef.Parts[0] == reIndexColName &&
					indexdef.IndexAlgo == tree.INDEX_TYPE_IVFFLAT.ToString() &&
					indexdef.IndexAlgoTableType == catalog.SystemSecondaryIndex_IvfCentroidsMappingRel {
					destIndexTableName = indexdef.IndexTableName
					found = true
					break
				}
			}
			if !found {
				return nil, moerr.NewInternalErrorNoCtx("aux2 table not found")
			}

			alterTableReIndex.Actions[0] = &plan.AlterTable_Action{
				Action: &plan.AlterTable_Action_ReindexRemapCol{
					ReindexRemapCol: &plan.AlterReindexRemapCol{
						DbName:               databaseName,
						TableName:            tableName,
						IndexSourceTableName: sourceIndexTableName,
						IndexDestTableName:   destIndexTableName,
						//IndexTableName:          indexTableName,
						//OriginTableIndexColName: reIndexColName,
						//OriginTableIndexColType: reIndexColType,
						IndexAlgo: reIndexAlgo.ToString(),
					},
				},
			}
		}

	}

	return &Plan{
		Plan: &plan.Plan_Ddl{
			Ddl: &plan.DataDefinition{
				DdlType: plan.DataDefinition_ALTER_TABLE,
				Definition: &plan.DataDefinition_AlterTable{
					AlterTable: alterTableReIndex,
				},
			},
		},
	}, nil

}
