package infrastructure

import (
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewDatabaseConnectionFromURL(url string) (*sqlx.DB, error) {
	driver := "pgx"
	if strings.HasPrefix(url, "mysql://") {
		driver = "mysql"
		// Convert mysql://user:pass@host:port/db to user:pass@tcp(host:port)/db
		url = strings.TrimPrefix(url, "mysql://")
		parts := strings.SplitN(url, "@", 2)
		if len(parts) == 2 {
			creds := parts[0]
			rest := parts[1]
			addrParts := strings.SplitN(rest, "/", 2)
			if len(addrParts) == 2 {
				addr := addrParts[0]
				dbName := addrParts[1]
				url = fmt.Sprintf("%s@tcp(%s)/%s", creds, addr, dbName)
			}
		}
	} else if strings.HasPrefix(url, "postgresql://") || strings.HasPrefix(url, "postgres://") {
		driver = "pgx"
	}

	return connect(driver, url)
}

func NewDatabaseConnection(driver, user, password, host, port, dbname string) (*sqlx.DB, error) {
	var dsn string
	actualDriver := "pgx"

	if driver == "mysql" {
		actualDriver = "mysql"
		addr := host
		if port != "" && !strings.Contains(host, ":") {
			addr = fmt.Sprintf("%s:%s", host, port)
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, password, addr, dbname)
	} else {
		// Default to Postgres
		addr := host
		if port != "" && !strings.Contains(host, ":") {
			addr = fmt.Sprintf("%s:%s", host, port)
		}
		dsn = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, addr, dbname)
	}

	return connect(actualDriver, dsn)
}

func connect(driver, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		return nil, err
	}

	// Connection pool settings
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	log.Printf("Connected to %s database", driver)
	return db, nil
}
