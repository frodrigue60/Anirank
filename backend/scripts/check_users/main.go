package main

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func main() {
	dsn := "root:@tcp(127.0.0.1:3306)/anirank?parseTime=true"
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	var users []struct {
		ID   uint64  `db:"id"`
		Name string  `db:"name"`
		Slug *string `db:"slug"`
	}

	err = db.Select(&users, "SELECT id, name, slug FROM users LIMIT 10")
	if err != nil {
		log.Fatalln(err)
	}

	for _, u := range users {
		slug := "NULL"
		if u.Slug != nil {
			slug = *u.Slug
		}
		fmt.Printf("ID: %d, Name: %s, Slug: %s\n", u.ID, u.Name, slug)
	}
}
