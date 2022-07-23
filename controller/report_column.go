package controller

import (
	"fmt"
	"repgen/core"
	"strings"
	"time"
)

type ReportColumn struct {
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

var ReportColumnTypeMap = map[int]struct{}{
	ReportColumnTypeStr:     emptyStruct,
	ReportColumnTypeInt:     emptyStruct,
	ReportColumnTypeFloat:   emptyStruct,
	ReportColumnTypeFormula: emptyStruct,
}

func CreateReportColumns(reportColumns []ReportColumn) error {
	columns := []string{"report_id", "name", "type", "formula", "created", "created_user_id"}
	sql := fmt.Sprintf("INSERT INTO report_column (%s) VALUES %s RETURNING id",
		strings.Join(columns, ","), core.PrepareQueryBulk(len(columns), len(reportColumns)))

	values := []interface{}{}
	for _, row := range reportColumns {
		values = append(values, row.ReportId, row.Name, row.Type, row.Formula, row.Created, row.CreatedUserId)
	}
	rows, err := core.Database.Query(sql, values...)
	if err != nil {
		return err
	}
	defer rows.Close()
	index := 0
	for rows.Next() {
		err := rows.Scan(&reportColumns[index].Id)
		if err != nil {
			return err
		}
		index++
	}
	return nil
}

func PopulateReportColumns(report *Report) error {
	rows, err := core.Database.Query("SELECT id, report_id, name, type, formula, created, created_user_id "+
		"FROM report_column WHERE report_id = $1", report.Id)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var reportColumn ReportColumn
		err := rows.Scan(&reportColumn.Id, &reportColumn.ReportId, &reportColumn.Name, &reportColumn.Type,
			&reportColumn.Formula, &reportColumn.Created, &reportColumn.CreatedUserId)
		if err != nil {
			return err
		}
		report.Columns = append(report.Columns, reportColumn)
	}
	return nil
}
