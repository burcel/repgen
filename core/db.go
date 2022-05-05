package core

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var Database *sql.DB

func InitializeDatabase() {
	// Create connection
	Database, err := sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%s/%s", Config.Postgresql.User,
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
