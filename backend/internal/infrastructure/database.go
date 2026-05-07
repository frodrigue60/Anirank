package infrastructure

import (
	"fmt"
	"log"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewDatabaseConnection(user, password, host, port, dbname string) (*sqlx.DB, error) {
	// If the host already contains a port (common in Railway), we don't append another one
	addr := host
	if port != "" && !strings.Contains(host, ":") {
		addr = fmt.Sprintf("%s:%s", host, port)
	}

	// Standard PostgreSQL DSN format
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		user, password, addr, dbname)

	return connect(dsn)
}

func NewDatabaseConnectionFromURL(url string) (*sqlx.DB, error) {
	return connect(url)
}

func connect(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Connection pool settings
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)

	log.Println("Connected to PostgreSQL database (via pgx)")
	return db, nil
}
