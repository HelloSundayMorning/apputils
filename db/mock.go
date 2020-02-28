package db

import (
	"database/sql"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type (
	MockDB struct {
		*sql.DB
	}

	TxMockDb struct {
		tx *sql.Tx
	}
)

func NewMockDB() (mockDb *MockDB, mock sqlmock.Sqlmock, err error) {

	sqlDb, mock, err := sqlmock.New()

	if err != nil {
		return nil, mock, err
	}

	mockDb = &MockDB{
		DB: sqlDb,
	}

	return mockDb, mock, nil


}

func (pDb *MockDB) GetDB() *sql.DB {
	return pDb.DB
}

func (pDb *MockDB) WithTx(txFunc func(tx AppSqlTx) error) (err error) {

	tx, err := pDb.Begin()

	if err != nil {
		return err
	}

	err = txFunc(&TxMockDb{tx})

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

func (txDb *TxMockDb) GetTx() *sql.Tx {
	return txDb.tx
}
