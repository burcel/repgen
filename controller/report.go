package controller

import (
	"repgen/core"
	"time"
)

type Report struct {
	Id            int
	ProjectId     int
	Name          string
	Interval      int
	Token         string
	Description   string
	Created       time.Time
	CreatedUserId int
	Columns       []ReportColumn
}

const (
	ReportNameMaxLength        = 200
	ReportTokenLength          = 20
	ReportDescriptionMaxLength = 1000
	ReportIntervalMonthly      = 0
	ReportIntervalWeekly       = 1
	ReportIntervalDaily        = 2
	ReportIntervalHourly       = 3
)

var emptyStruct struct{}
var ReportIntervalMap = map[int]struct{}{
	ReportIntervalMonthly: emptyStruct,
	ReportIntervalWeekly:  emptyStruct,
	ReportIntervalDaily:   emptyStruct,
	ReportIntervalHourly:  emptyStruct,
}

func CreateReport(report *Report) error {
	rows, err := core.Database.Query("INSERT INTO report (project_id, name, interval, token, description, created, created_user_id) "+
		"VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id", report.ProjectId, report.Name, report.Interval, report.Token,
		report.Description, report.Created, report.CreatedUserId)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&report.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetReportByToken(token string) (report *Report, err error) {
	rows, err := core.Database.Query("SELECT id, project_id, name, interval, token, description, created, created_user_id "+
		"FROM report WHERE token = $1", token)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		report = &Report{}
		err := rows.Scan(&report.Id, &report.ProjectId, &report.Name, &report.Interval, &report.Token,
			&report.Description, &report.Created, &report.CreatedUserId)
		if err != nil {
			return nil, err
		}
	}
	return report, nil
}
