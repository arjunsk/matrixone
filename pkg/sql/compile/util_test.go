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

package compile

import (
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"testing"
)

func Test_genCreateIndexTableSql(t *testing.T) {
	type args struct {
		indexTableDef *plan.TableDef
		indexDef      *plan.IndexDef
		DBName        string
		isUK          bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test UK",
			args: args{
				indexTableDef: &plan.TableDef{
					Name: "t1",
					Cols: []*plan.ColDef{
						{
							Name: "c1",
							Typ: &plan.Type{
								Id: int32(types.T_int32),
							},
						},
						{
							Name: "c2",
							Typ: &plan.Type{
								Id: int32(types.T_int64),
							},
						},
					},
				},
				indexDef: &plan.IndexDef{
					IndexTableName: "tbl1",
				},
				DBName: "db1",
				isUK:   true,
			},
			want: "create table db1.`tbl1` (c1 INT primary key,c2 BIGINT);",
		},
		{
			name: "Test SK",
			args: args{
				indexTableDef: &plan.TableDef{
					Name: "t1",
					Cols: []*plan.ColDef{
						{
							Name: "c1",
							Typ: &plan.Type{
								Id: int32(types.T_int32),
							},
						},
						{
							Name: "c2",
							Typ: &plan.Type{
								Id: int32(types.T_int64),
							},
						},
					},
				},
				indexDef: &plan.IndexDef{
					IndexTableName: "tbl1",
				},
				DBName: "db1",
				isUK:   false,
			},
			want: "create table db1.`tbl1` (c1 INT,c2 BIGINT) CLUSTER BY(c1);",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genCreateIndexTableSql(tt.args.indexTableDef, tt.args.indexDef, tt.args.DBName, tt.args.isUK); got != tt.want {
				t.Errorf("genCreateIndexTableSql() = %v, want %v", got, tt.want)
			}
		})
	}
}
