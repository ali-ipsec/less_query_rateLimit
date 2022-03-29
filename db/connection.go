package db

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "postgres"
)

var doOnce sync.Once
var singleton *sql.DB

func GetConnection() *sql.DB {
	doOnce.Do(func() {
		conninfo := fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
		db, _ := sql.Open("postgres", conninfo)
		singleton = db
	})
	return singleton
}
