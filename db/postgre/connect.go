package Database

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func Connect() {
	db, err := sql.Open("postgres", connStr)
    if err != nil {
        panic(err.Error())
    }
    defer db.Close()
}
