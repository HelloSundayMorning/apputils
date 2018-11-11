package db

import (
	"database/sql"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type (
	MockDB struct {
		*sql.DB
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

func (pDb *MockDB) GetDB() (*sql.DB) {
	return pDb.DB
}
