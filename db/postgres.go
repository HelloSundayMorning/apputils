package db

import (
	"database/sql"
	"fmt"
	"github.com/HelloSundayMorning/apputils/appctx"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
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

func NewPostgresDBWithPort(host, user, pw, dbName, port string) (pgDb *PostgresDB, err error) {

	dataSource := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, pw, dbName, port)

	db, err := sql.Open("postgres", dataSource)

	if err != nil {
		return pgDb, err
	}

	pgDb = &PostgresDB{
		DB: db,
	}

	return pgDb, nil

}

func (pDb *PostgresDB) GetDB() *sql.DB {
	return pDb.DB
}

func (pDb *PostgresDB) SqlTxFromContext(ctx context.Context) (tx *sql.Tx, err error) {

	store := ctx.(appctx.AppContext).ValueStore

	val, ok := store[appctx.SqlTransactionKey]

	if !ok {
		tx, err = pDb.BeginTx(ctx, nil)

		if err != nil {
			return tx, err
		}

		store[appctx.SqlTransactionKey] = tx
	} else {
		tx = val.(*sql.Tx)
	}

	return tx, nil
}

