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
