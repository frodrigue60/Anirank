package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	res, err := db.Exec("DELETE FROM images WHERE imageable_type = 'App\\\\Models\\\\User' AND imageable_id = 1")
	if err != nil {
		log.Fatalln(err)
	}

	affected, _ := res.RowsAffected()
	fmt.Printf("Deleted %d duplicate rows for User 1.\n", affected)
}
