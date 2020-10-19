package db

import (
	"database/sql"
	"errors"
	"github.com/HelloSundayMorning/apputils/log"
	"golang.org/x/net/context"
	"strings"
)

var ErrMigratingColumnAlreadyExists = errors.New("migrating: column already exists")

func (pDb *PostgresDB) Migrate(ctx context.Context, addColumnStatements []string) (err error) {

	component := "migrate"

	log.Printf(ctx, component, "Starting Migration")

	conn := pDb.GetDB()

	for _, m := range addColumnStatements {

		log.Printf(ctx, component, "Migrating %s", m)

		err = pDb.migrateAddColumn(ctx, conn, m)

		if err != nil && err != ErrMigratingColumnAlreadyExists {
			return err
		}

		if err == ErrMigratingColumnAlreadyExists {
			continue
		}
	}

	log.Printf(ctx, component, "Migration Completed")

	return nil

}

func (pDb *PostgresDB) migrateAddColumn(ctx context.Context, conn *sql.DB, addColumnStatement string) (err error) {

	tx, err := conn.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(addColumnStatement)

	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = stmt.Exec()

	if err != nil {
		if strings.HasSuffix(err.Error(), "already exists") {
			log.Printf(ctx, "MigrateAddColumn", "Migrating column: %s", err)
			return ErrMigratingColumnAlreadyExists
		}

		log.Errorf(ctx, "MigrateAddColumn", "Error migrating column: %s, %s", addColumnStatement, err)
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}

