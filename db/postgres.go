package db

import (
	"database/sql"
	"fmt"
	"github.com/aws/aws-xray-sdk-go/xray"
	_ "github.com/lib/pq"
)

type (
	PostgresDB struct {
		*sql.DB
	}

	TxPostgresDb struct {
		tx *sql.Tx
	}
)

func (txDb *TxPostgresDb) GetTx() *sql.Tx {
	return txDb.tx
}

func NewPostgresDB(host, user, pw, dbName string) (pgDb *PostgresDB, err error) {

	dataSource := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, pw, dbName)

	// Here we Open connection to the database using AWS XRay SQL context. Same as sql.Open but wrapped in the XRay lib
	db, err := xray.SQLContext("postgres", dataSource)

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

	// Here we Open connection to the database using AWS XRay SQL context. Same as sql.Open but wrapped in the XRay lib
	db, err := xray.SQLContext("postgres", dataSource)

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


func (pDb *PostgresDB) WithTx(txFunc func(tx AppSqlTx) error) (err error) {

	tx, err := pDb.Begin()

	if err != nil {
		return err
	}

	err = txFunc(&TxPostgresDb{tx})

	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()

	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}
