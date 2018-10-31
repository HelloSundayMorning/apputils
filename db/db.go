package db

import "database/sql"

type (
	AppSqlDb interface {
		GetDB() (*sql.DB)
	}
)
