package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)


type (

	PostgresDB struct {
		*sql.DB
	}
)

func NewPostgresDB(host, user, pw, dbName string) (pgDb *PostgresDB, err error) {

	dataSource := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, pw, dbName)

	db, err := sql.Open("postgres", dataSource)

	if err != nil {
		return pgDb, err
	}

	pgDb = &PostgresDB{
		DB: db,
	}

	return pgDb, nil

}

func (pDb *PostgresDB) GetDB() (*sql.DB) {
	return pDb.DB
}
