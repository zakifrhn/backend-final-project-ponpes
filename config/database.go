package config

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/lib/pq"
)

const (
    host     = "154.19.37.208"
    port     = 5432
    user     = "alaziziyah"
    password = "p0npes123"
    dbname   = "ponpes_alaziziyah"
)

func ConnectDB() *sql.DB {
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)

    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        log.Fatal(err)
    }

    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Successfully connected to database!")
    return db
}