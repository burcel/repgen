package controller

import (
	"fmt"
	"repgen/core"
	"time"
)

type ReportDefinition struct {
	Id            int
	ReportId      int
	Name          string
	Type          int
	Formula       string
	Created       time.Time
	CreatedUserId int
}

const (
	ReportColumnMaxCount         = 30
	ReportColumnNameMaxLength    = 100
	ReportColumnTypeStr          = 0
	ReportColumnTypeInt          = 1
	ReportColumnTypeFloat        = 2
	ReportColumnTypeFormula      = 3
	ReportColumnFormulaMaxLength = 200
)

var ReportDefinitionTypeMap = map[int]struct{}{
	ReportColumnTypeStr:     emptyStruct,
	ReportColumnTypeInt:     emptyStruct,
	ReportColumnTypeFloat:   emptyStruct,
	ReportColumnTypeFormula: emptyStruct,
}

func CreateReportDefinition(reportDefinition []ReportDefinition) error {
	sql := fmt.Sprintf("INSERT INTO report_definition (report_id, name, type, formula, created, created_user_id) "+
		"VALUES %s", core.PrepareQueryBulk(6, len(reportDefinition)))
	stmt, err := core.Database.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	values := []interface{}{}
	for _, row := range reportDefinition {
		values = append(values, row.ReportId, row.Name, row.Type, row.Formula, row.Created, row.CreatedUserId)
	}
	_, err = stmt.Exec(values...)
	if err != nil {
		return err
	}
	return nil
}
