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
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/sql/parsers/tree"
)

func buildAlterTableReindex(stmt *tree.AlterTable, ctx CompilerContext) (*Plan, error) {
	if len(stmt.Options) != 1 {
		return nil, moerr.NewInternalErrorNoCtx("currently we only support reindexing one column")
	}

	tableName := string(stmt.Table.ObjectName)
	databaseName := string(stmt.Table.SchemaName)
	if databaseName == "" {
		databaseName = ctx.DefaultDatabase()
	}
	_, tableDef := ctx.Resolve(databaseName, tableName)
	if tableDef == nil {
		return nil, moerr.NewNoSuchTable(ctx.GetContext(), databaseName, tableName)
	}

	alterTable := &plan.AlterTable{
		Actions:       make([]*plan.AlterTable_Action, len(stmt.Options)),
		AlgorithmType: plan.AlterTable_DEFAULT, //TODO: change to DEFAULT
	}

	oriPriKeyName := getTablePriKeyName(tableDef.Pkey)

	_, _, secKeyName := stmt.Options[0].(*tree.AlterOptionReindex).ColumnName.GetNames()
	_, secKeyType := getSecKeyPos(tableDef, secKeyName)

	// check index
	found := false
	var indexTableName string
	for _, indexdef := range tableDef.Indexes {
		if secKeyName == indexdef.IndexName {
			indexTableName = indexdef.IndexTableName
			found = true
			break
		}
	}

	if !found {
		return nil, moerr.NewInternalErrorNoCtx("not found")
	}

	alterTable.Actions[0] = &plan.AlterTable_Action{
		Action: &plan.AlterTable_Action_ReindexCol{
			ReindexCol: &plan.AlterReindexCol{
				DbName:                      databaseName,
				TableName:                   tableName,
				OriginTablePrimaryKeyName:   oriPriKeyName,
				OriginTableSecondaryKeyName: secKeyName,
				OriginTableSecondaryKeyType: secKeyType,
				IndexTableExist:             true,
				IndexTableName:              indexTableName,
			},
		},
	}

	return &Plan{
		Plan: &plan.Plan_Ddl{
			Ddl: &plan.DataDefinition{
				DdlType: plan.DataDefinition_ALTER_TABLE,
				Definition: &plan.DataDefinition_AlterTable{
					AlterTable: alterTable,
				},
			},
		},
	}, nil
}
