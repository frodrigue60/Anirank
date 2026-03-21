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
	godotenv.Load(".env")
	dbURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := sqlx.Connect("mysql", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Update thumbnail_src
	res, err := db.Exec("UPDATE animes a JOIN images i ON a.id = i.imageable_id AND i.imageable_type = 'App\\\\Models\\\\Post' AND i.type = 'thumbnail' SET a.thumbnail_src = i.path")
	if err != nil {
		log.Fatal("Thumbnail error:", err)
	}
	rows, _ := res.RowsAffected()
	fmt.Printf("Thumbnail paths synced: %d rows\n", rows)

	// Update banner_src
	res2, err2 := db.Exec("UPDATE animes a JOIN images i ON a.id = i.imageable_id AND i.imageable_type = 'App\\\\Models\\\\Post' AND i.type = 'banner' SET a.banner_src = i.path")
	if err2 != nil {
		log.Fatal("Banner error:", err2)
	}
	rows2, _ := res2.RowsAffected()
	fmt.Printf("Banner paths synced: %d rows\n", rows2)
}
