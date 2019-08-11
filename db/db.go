package db

import (
	"database/sql"
)

type (
	AppSqlDb interface {
		GetDB() *sql.DB
		WithTx(txFunc func(tx AppSqlTx) error) (err error)
	}

	AppSqlTx interface {
		GetTx() *sql.Tx

	}

	AppRepository interface {
		WithTx(txFunc func(tx AppSqlTx) error) (err error)
	}


)
