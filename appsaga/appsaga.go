package appsaga

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/HelloSundayMorning/apputils/app"
	"github.com/HelloSundayMorning/apputils/db"
	"github.com/HelloSundayMorning/apputils/eventpubsub"
	"github.com/HelloSundayMorning/apputils/log"
	"github.com/lib/pq"
	"golang.org/x/net/context"
	"time"
)

type (
	Saga struct {
		SagaName   string
		SagaKey    string
		Events     map[string]eventpubsub.AppEvent
		EventTypes []string
		Completed  bool
		Timestamp  int64
	}

	ISagaManager interface {
		AddEvent(ctx context.Context, sagaKey string, appEvent eventpubsub.AppEvent) (saga Saga, err error)
	}

	SagaManager struct {
		AppID            app.ApplicationID
		sqlDb            db.AppSqlDb
		SagaName         string
		EventTypes       []string
		completedHandler SagaCompletedHandler
	}

	SagaCompletedHandler func(ctx context.Context, saga Saga) (err error)
)

const (
	createSagaTable = `CREATE TABLE IF NOT EXISTS saga_manager (
									 saga_name                  varchar(50)               not null,
									 saga_key                   varchar(500)              not null,
                                     timestamp                  bigint                    not null,
                                     events                     json default '{}' :: json not null,
                                     event_types                text []                   not null,
                                     completed                  boolean                   not null,
                                     PRIMARY KEY (saga_name, saga_key));`

	insertSaga = `INSERT INTO saga_manager (saga_name, saga_key, timestamp, events , event_types, completed)
                                VALUES ($1, $2, $3, $4, $5, $6)
                                ON CONFLICT (saga_name, saga_key) DO UPDATE SET
                                timestamp = $3,
                                events = $4,
                                event_types = $5,
                                completed = $6`

	findSaga = `SELECT saga_name, saga_key, timestamp, events , event_types, completed
                    FROM saga_manager
                    WHERE saga_name = $1 AND saga_key = $2`
)

func NewSagaManager(appID app.ApplicationID, sqlDb db.AppSqlDb, sagaName string, eventTypes []string, sagaCompletedHandler SagaCompletedHandler) (manager *SagaManager, err error) {

	manager = &SagaManager{
		AppID:            appID,
		sqlDb:            sqlDb,
		SagaName:         sagaName,
		EventTypes:       eventTypes,
		completedHandler: sagaCompletedHandler,
	}

	err = manager.initialize()

	if err != nil {
		return nil, err
	}

	log.PrintfNoContext(manager.AppID, "sagaManager", "Saga Manager %s for eventTypes: %s", sagaName, eventTypes)

	return manager, nil
}

func (sagaManager *SagaManager) initialize() (err error) {

	conn := sagaManager.sqlDb.GetDB()

	ctx := context.Background()

	tx, err := conn.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(createSagaTable)

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = stmt.Exec()

	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}

func (sagaManager *SagaManager) store(ctx context.Context, tx *sql.Tx, saga *Saga) (err error) {

	stmt, err := tx.Prepare(insertSaga)

	if err != nil {
		tx.Rollback()
		return err
	}

	eventsMapJSON, err := json.Marshal(saga.Events)

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = stmt.Exec(saga.SagaName, saga.SagaKey, saga.Timestamp, eventsMapJSON, pq.Array(saga.EventTypes), saga.Completed)

	if err != nil {
		tx.Rollback()
		return err
	}

	return nil

}

func (sagaManager *SagaManager) findSagaByKey(ctx context.Context, tx *sql.Tx, sagaKey string) (userSaga Saga, err error) {

	rows, err := tx.Query(findSaga, sagaManager.SagaName, sagaKey)

	if err != nil {
		return userSaga, err
	}

	defer rows.Close()

	userSaga.Events = make(map[string]eventpubsub.AppEvent)

	var eventsMapJSON json.RawMessage

	if rows.Next() {

		err = rows.Scan(&userSaga.SagaName, &userSaga.SagaKey, &userSaga.Timestamp, &eventsMapJSON, pq.Array(&userSaga.EventTypes), &userSaga.Completed)

		if err != nil {
			return userSaga, err
		}

		err = json.Unmarshal(eventsMapJSON, &userSaga.Events)

		if err != nil {
			return userSaga, err
		}

	}

	return userSaga, nil

}

func (sagaManager *SagaManager) AddEvent(ctx context.Context, sagaKey string, appEvent eventpubsub.AppEvent) (saga Saga, err error) {

	if !sagaManager.isValidEvent(appEvent) {
		return saga, fmt.Errorf("invalid event type %s add attempt to saga %s", appEvent.EventType, sagaManager.SagaName)
	}

	conn := sagaManager.sqlDb.GetDB()

	tx, err := conn.Begin()

	if err != nil {
		tx.Rollback()
		return saga, err
	}

	_, err = tx.Exec(`set transaction isolation level serializable`)   // <=== SET ISOLATION LEVEL

	if err != nil {
		tx.Rollback()
		return saga, err
	}

	saga, err = sagaManager.findSagaByKey(ctx, tx, sagaKey)

	if err != nil {
		tx.Rollback()
		return saga, err
	}

	if saga.Completed {
		err = tx.Commit()

		if err != nil {
			return saga, err
		}

		log.Printf(ctx, "sagaManager", "Saga %s key %s was completed previously", saga.SagaName, saga.SagaKey)
		return saga, nil
	}

	log.Printf(ctx, "sagaManager", "SLEEPING")
	time.Sleep(30* time.Second)
	log.Printf(ctx, "sagaManager", "WAKEUP")

	saga.SagaName = sagaManager.SagaName
	saga.EventTypes = sagaManager.EventTypes
	saga.SagaKey = sagaKey
	saga.Timestamp = time.Now().UTC().UnixNano()
	saga.Events[appEvent.EventType] = appEvent

	saga.Completed = sagaManager.validateCompleted(saga)

	err = sagaManager.store(ctx, tx, &saga)

	err = tx.Commit()

	if err != nil {
		return saga, err
	}

	log.Printf(ctx, "sagaManager", "Saga %s key %s updated with event type %s. Completed %t", saga.SagaName, saga.SagaKey, appEvent.EventType, saga.Completed)

	if saga.Completed {

		log.Printf(ctx, "sagaManager", "Saga %s key %s Completed. Handling saga..., %s", saga.SagaName, saga.SagaKey, err)

		err = sagaManager.completedHandler(ctx, saga)

		if err != nil {

			log.Printf(ctx, "sagaManager", "Saga %s key %s Completed. But got error while executing completed handler. Completion rolled back..., %s", saga.SagaName, saga.SagaKey, err)

			tx, err := conn.Begin()

			saga.Completed = false

			err = sagaManager.store(ctx, tx, &saga)

			err = tx.Commit()

			if err != nil {
				return saga, err
			}

			return saga, err
		}

	}


	return saga, nil
}

func (sagaManager *SagaManager) isValidEvent(appEvent eventpubsub.AppEvent) bool {

	for _, eventType := range sagaManager.EventTypes {

		if eventType == appEvent.EventType {
			return true
		}

	}

	return false
}

func (sagaManager *SagaManager) validateCompleted(saga Saga) bool {

	if len(sagaManager.EventTypes) == len(saga.Events) {
		return true
	}

	return false
}


