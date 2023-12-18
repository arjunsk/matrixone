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
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
)

var (
	bigIntType = types.T_int64.ToType()
)

func makeMetaTblScanWhereKeyEqVersion(builder *QueryBuilder, bindCtx *BindContext, indexTableDefs []*TableDef, idxRefs []*ObjectRef) (int32, error) {
	var metaTableScanId int32

	scanNodeProject := make([]*Expr, len(indexTableDefs[0].Cols))
	for colIdx, column := range indexTableDefs[0].Cols {
		scanNodeProject[colIdx] = &plan.Expr{
			Typ: column.Typ,
			Expr: &plan.Expr_Col{
				Col: &plan.ColRef{
					ColPos: int32(colIdx),
					Name:   column.Name,
				},
			},
		}
	}
	metaTableScanId = builder.appendNode(&Node{
		NodeType:    plan.Node_TABLE_SCAN,
		Stats:       &plan.Stats{},
		ObjRef:      idxRefs[0],
		TableDef:    indexTableDefs[0],
		ProjectList: scanNodeProject,
	}, bindCtx)

	prevProjection := getProjectionByLastNode(builder, metaTableScanId)
	conditionExpr, err := BindFuncExprImplByPlanExpr(builder.GetContext(), "=", []*Expr{
		prevProjection[0],
		MakePlan2StringConstExprWithType("version"),
	})
	if err != nil {
		return -1, err
	}
	metaTableScanId = builder.appendNode(&Node{
		NodeType:   plan.Node_FILTER,
		Children:   []int32{metaTableScanId},
		FilterList: []*Expr{conditionExpr},
	}, bindCtx)

	prevProjection = getProjectionByLastNode(builder, metaTableScanId)
	castValueCol, err := makePlan2CastExpr(builder.GetContext(), prevProjection[1], makePlan2Type(&bigIntType))
	if err != nil {
		return -1, err
	}
	metaTableScanId = builder.appendNode(&Node{
		NodeType:    plan.Node_PROJECT,
		Stats:       &plan.Stats{},
		Children:    []int32{metaTableScanId},
		ProjectList: []*Expr{castValueCol},
	}, bindCtx)

	return metaTableScanId, nil
}

func makeCentroidsTblScan(builder *QueryBuilder, bindCtx *BindContext, indexTableDefs []*TableDef, idxRefs []*ObjectRef) int32 {
	scanNodeProject := make([]*Expr, len(indexTableDefs[1].Cols))
	for colIdx, column := range indexTableDefs[1].Cols {
		scanNodeProject[colIdx] = &plan.Expr{
			Typ: column.Typ,
			Expr: &plan.Expr_Col{
				Col: &plan.ColRef{
					ColPos: int32(colIdx),
					Name:   column.Name,
				},
			},
		}
	}
	centroidsScanId := builder.appendNode(&Node{
		NodeType:    plan.Node_TABLE_SCAN,
		Stats:       &plan.Stats{},
		ObjRef:      idxRefs[1],
		TableDef:    indexTableDefs[1],
		ProjectList: scanNodeProject,
	}, bindCtx)
	return centroidsScanId
}

func makeCrossJoinCentroidsMetaForCurrVersion(builder *QueryBuilder, bindCtx *BindContext, indexTableDefs []*TableDef, idxRefs []*ObjectRef, metaTableScanId int32) (int32, error) {
	centroidsScanId := makeCentroidsTblScan(builder, bindCtx, indexTableDefs, idxRefs)

	joinProjections := getProjectionByLastNode(builder, centroidsScanId)[:3]
	joinProjections = append(joinProjections, &plan.Expr{
		Typ: makePlan2Type(&bigIntType),
		Expr: &plan.Expr_Col{
			Col: &plan.ColRef{
				RelPos: 1,
				ColPos: 0,
			},
		},
	})
	joinMetaAndCentroidsId := builder.appendNode(&plan.Node{
		NodeType:    plan.Node_JOIN,
		JoinType:    plan.Node_SINGLE,
		Children:    []int32{centroidsScanId, metaTableScanId},
		ProjectList: joinProjections,
	}, bindCtx)

	prevProjection := getProjectionByLastNode(builder, joinMetaAndCentroidsId)
	conditionExpr, err := BindFuncExprImplByPlanExpr(builder.GetContext(), "=", []*Expr{
		prevProjection[0],
		prevProjection[3],
	})
	if err != nil {
		return -1, err
	}
	filterId := builder.appendNode(&plan.Node{
		NodeType:    plan.Node_FILTER,
		Children:    []int32{joinMetaAndCentroidsId},
		FilterList:  []*Expr{conditionExpr},
		ProjectList: prevProjection[:3],
	}, bindCtx)
	return filterId, nil
}

func makeCrossJoinTblAndCentroids(builder *QueryBuilder, bindCtx *BindContext, tableDef *TableDef,
	leftChildTblId int32, rightChildCentroidsId int32,
	typeOriginPk *Type, posOriginPk int,
	typeOriginVecColumn *Type, posOriginVecColumn int) int32 {

	crossJoinID := builder.appendNode(&plan.Node{
		NodeType: plan.Node_JOIN,
		JoinType: plan.Node_INNER,
		Children: []int32{leftChildTblId, rightChildCentroidsId},
		ProjectList: []*Expr{
			{
				// centroids.version
				Typ: makePlan2Type(&bigIntType),
				Expr: &plan.Expr_Col{
					Col: &plan.ColRef{
						RelPos: 1,
						ColPos: 0,
						Name:   catalog.SystemSI_IVFFLAT_TblCol_Centroids_version,
					},
				},
			},
			{ // centroids.centroid_id
				Typ: makePlan2Type(&bigIntType),
				Expr: &plan.Expr_Col{
					Col: &plan.ColRef{
						RelPos: 1,
						ColPos: 1,
						Name:   catalog.SystemSI_IVFFLAT_TblCol_Centroids_id,
					},
				},
			},
			{ // tbl.pk
				Typ: typeOriginPk,
				Expr: &plan.Expr_Col{
					Col: &plan.ColRef{
						RelPos: 0,
						ColPos: int32(posOriginPk),
						Name:   tableDef.Cols[posOriginPk].Name,
					},
				},
			},
			{ // centroids.centroid
				Typ: typeOriginVecColumn,
				Expr: &plan.Expr_Col{
					Col: &plan.ColRef{
						RelPos: 1,
						ColPos: 2,
						Name:   catalog.SystemSI_IVFFLAT_TblCol_Centroids_centroid,
					},
				},
			},
			{ // tbl.embedding
				Typ: typeOriginVecColumn,
				Expr: &plan.Expr_Col{
					Col: &plan.ColRef{
						RelPos: 0,
						ColPos: int32(posOriginVecColumn),
						Name:   tableDef.Cols[posOriginVecColumn].Name,
					},
				},
			},
		},
	}, bindCtx)

	return crossJoinID
}

func makeSortByL2DistAndLimit1AndProject4(builder *QueryBuilder, bindCtx *BindContext,
	crossJoinTblAndCentroidsID int32) (int32, error) {

	// 0: centroids.version,
	// 1: centroids.centroid_id,
	// 2: tbl.pk,
	// 3: centroids.centroid,
	// 4: tbl.embedding
	var joinProjections = getProjectionByLastNode(builder, crossJoinTblAndCentroidsID)
	cpKeyCol, err := BindFuncExprImplByPlanExpr(builder.GetContext(), "serial", []*plan.Expr{joinProjections[0], joinProjections[2]})
	if err != nil {
		return -1, err
	}

	l2Distance, err := BindFuncExprImplByPlanExpr(builder.GetContext(), "l2_distance", []*Expr{joinProjections[3], joinProjections[4]})
	if err != nil {
		return -1, err
	}

	sortId := builder.appendNode(&plan.Node{
		NodeType: plan.Node_SORT,
		Children: []int32{crossJoinTblAndCentroidsID},
		// version, centroid_id, pk, serial(version,pk)
		ProjectList: []*Expr{joinProjections[0], joinProjections[1], joinProjections[2], cpKeyCol},
		OrderBy: []*plan.OrderBySpec{
			{
				Flag: plan.OrderBySpec_ASC,
				Expr: l2Distance,
			},
		},
		Limit: makePlan2Int64ConstExprWithType(1),
	}, bindCtx)

	return sortId, nil
}
