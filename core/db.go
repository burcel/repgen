package core

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var Database *sql.DB

func InitializeDatabase() {
	// Create connection
	var err error
	Database, err = sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%s/%s", Config.Postgresql.User,
		Config.Postgresql.Password, Config.Postgresql.Host, Config.Postgresql.Port, Config.Postgresql.Database))
	if err != nil {
		panic(err)
	}
	// Check connection
	if err := Database.Ping(); err != nil {
		panic(err)
	}
	log.Println("Successfully connected to the PostgreSQL database.")

	// Maximum Idle Connections
	Database.SetMaxIdleConns(Config.Postgresql.MaxIdleConnections)
	// Maximum Open Connections
	Database.SetMaxOpenConns(Config.Postgresql.MaxOpenConnections)
	// Idle Connection Timeout
	Database.SetConnMaxIdleTime(1 * time.Minute)
	// Connection Lifetime
	Database.SetConnMaxLifetime(5 * time.Minute)
}

// columnCount: How many column value e.g. $1,$2 etc
// valueCount: How many values e.g. (...), (...)
// PrepareQueryBulk(3, 1) -> ($1,$2,$3)
// PrepareQueryBulk(3, 2) -> ($1,$2,$3),($4,$5,$6)
func PrepareQueryBulk(columnCount int, valueCount int) string {
	index := 1
	values := make([]string, valueCount)
	for i := 0; i < valueCount; i++ {
		columns := make([]string, columnCount)
		for j := 0; j < columnCount; j++ {
			columns[j] = fmt.Sprintf("$%d", index)
			index++
		}
		values[i] = fmt.Sprintf("(%s)", strings.Join(columns, ","))
	}
	return strings.Join(values, ",")
}

// columns: column names
// startingIndex: starting index for parameter
// PrepareQueryBulkUpdate(["x1", "x2"], 3) -> x1=$3,x2=$4
// PrepareQueryBulkUpdate(["a", "b", "c"], 7) -> a=$7,b=$8,c=$9
func PrepareQueryBulkUpdate(columns []string, startingIndex int) string {
	columnSql := make([]string, len(columns))
	for index, value := range columns {
		columnSql[index] = fmt.Sprintf("%s=$%d", value, index+startingIndex)
	}
	return strings.Join(columnSql, ",")
}
