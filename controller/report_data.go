package controller

import (
	"fmt"
	"repgen/core"
	"strings"
	"time"
)

type ReportData struct {
	Id         int
	ReportDate time.Time
	SentDate   time.Time
	ColumnMap  map[int]interface{}
}

const (
	ReportTableNamePattern  = "zz_report_%d"
	ReportColumnNamePattern = "C%d"
)

func ReturnReportTableName(reportId int) string {
	return fmt.Sprintf(ReportTableNamePattern, reportId)
}

func ReturnReportColumnName(reportColumnId int) string {
	return fmt.Sprintf(ReportColumnNamePattern, reportColumnId)
}

func ReturnReportColumnCreationSql(columns []ReportColumn) string {
	var sb strings.Builder
	for _, column := range columns {
		reportColumnName := ReturnReportColumnName(column.Id)
		switch column.Type {
		case ReportColumnTypeStr:
			sb.WriteString(fmt.Sprintf("%s varchar,\n", reportColumnName))
		case ReportColumnTypeInt:
			sb.WriteString(fmt.Sprintf("%s int,\n", reportColumnName))
		case ReportColumnTypeFloat:
			sb.WriteString(fmt.Sprintf("%s float,\n", reportColumnName))
		case ReportColumnTypeFormula:

		default:
			panic(fmt.Sprintf("Invalid report column type: %d", column.Type))
		}
	}
	return sb.String()
}

func CreateReportData(report Report) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()

	tableName := ReturnReportTableName(report.Id)
	sql := fmt.Sprintf(
		`CREATE TABLE %s (
			id int NOT NULL GENERATED ALWAYS AS IDENTITY,
			report_date timestamp without time zone NOT NULL,
			%s
			CONSTRAINT %s_pk PRIMARY KEY (id)
		)`,
		tableName, ReturnReportColumnCreationSql(report.Columns), tableName)
	stmt, err := core.Database.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	// Index
	sql = fmt.Sprintf("CREATE INDEX %s_idx ON %s USING brin(report_date)", tableName, tableName)
	stmt, err = core.Database.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}

func InsertReportData(reportData *ReportData) error {
	return nil
}
