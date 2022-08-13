package controller

import (
	"fmt"
	"repgen/core"
	"strconv"
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
	ReportTableNamePattern  = "zz_report_%d" // Report data tables should go at the end in table list
	ReportColumnNamePattern = "C%d"
)

func ReturnReportTableName(reportId int) string {
	return fmt.Sprintf(ReportTableNamePattern, reportId)
}

func ReturnReportColumnName(reportColumnId int) string {
	return fmt.Sprintf(ReportColumnNamePattern, reportColumnId)
}

func ReturnColumnId(reportColumnName string) (int, error) {
	columnId, err := strconv.Atoi(reportColumnName[1:])
	if err != nil {
		return 0, err
	}
	return columnId, nil
}

func returnReportColumnCreationSql(columns []ReportColumn) string {
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

func CreateReportDataTable(report Report) (err error) {
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
			sent_date timestamp without time zone NOT NULL,
			%s
			CONSTRAINT %s_pk PRIMARY KEY (id)
		)`,
		tableName, returnReportColumnCreationSql(report.Columns), tableName)
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

// TODO: report data should be updated if report dates coincide
func InsertReportData(reportId int, reportData *ReportData) error {
	// Prepare query columns and values
	columns := []string{"report_date", "sent_date"}
	values := []interface{}{reportData.ReportDate, reportData.SentDate}
	for key, value := range reportData.ColumnMap {
		columns = append(columns, ReturnReportColumnName(key))
		values = append(values, value)
	}
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s RETURNING id",
		ReturnReportTableName(reportId), strings.Join(columns, ","), core.PrepareQueryBulk(len(columns), 1))
	rows, err := core.Database.Query(sql, values...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&reportData.Id)
		if err != nil {
			return err
		}
	}
	return nil
}
