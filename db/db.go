package db

import (
	"database/sql"
	"golang.org/x/net/context"
)

type (
	AppSqlDb interface {
		GetDB() *sql.DB
		SqlTxFromContext(ctx context.Context) (tx *sql.Tx, err error)
	}
)
