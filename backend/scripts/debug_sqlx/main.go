package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
)

type SongTest struct {
	ID     uint64 `db:"id"`
	Status bool   `db:"status"`
}

func main() {
	dsn := "root:@tcp(127.0.0.1:3306)/anirank?parseTime=true"
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	var songs []SongTest
	err = db.Select(&songs, "SELECT s.* FROM songs s LIMIT 1")
	if err != nil {
		fmt.Printf("ERROR WITH s.*: %v\n", err)
	} else {
		fmt.Printf("SUCCESS WITH s.*: %+v\n", songs[0])
	}

	err = db.Select(&songs, "SELECT s.id, s.status FROM songs s JOIN animes a ON s.anime_id = a.id LIMIT 1")
	if err != nil {
		fmt.Printf("ERROR WITH JOIN: %v\n", err)
	} else {
		fmt.Printf("SUCCESS WITH JOIN: %+v\n", songs[0])
	}
}
