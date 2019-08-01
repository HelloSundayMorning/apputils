package appctx

import (
	"database/sql"
	"github.com/HelloSundayMorning/apputils/app"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
	"net/http"
	"reflect"
	"time"
)

const (
	CorrelationIdHeader    = "x-correlation-id"
	AppIdHeader            = "x-app-id"
	FromAppIdHeader        = "x-from-app-id"
	AuthorizedUserIDHeader = "x-authorized-user-id"
	SqlTransactionKey      = "x-sql-transaction-key"
)

type (
	AppContext struct {
		ValueStore map[string]interface{}
	}
)

func (ctx AppContext) Deadline() (deadline time.Time, ok bool) {
	return deadline, ok
}

func (ctx AppContext) Done() <-chan struct{} {

	cDone := make(chan struct{})

	return cDone
}

func (ctx AppContext) Err() error {
	return nil
}

func (ctx AppContext) Value(key interface{}) interface{} {

	if reflect.TypeOf(key) != reflect.TypeOf("") {
		return nil
	}

	return ctx.ValueStore[key.(string)]
}

func NewContextFromDelivery(appID app.ApplicationID, delivery amqp.Delivery) (ctx context.Context) {

	store := make(map[string]interface{})

	store[CorrelationIdHeader] = delivery.CorrelationId
	store[AppIdHeader] = string(appID)
	store[FromAppIdHeader] = delivery.AppId
	store[AuthorizedUserIDHeader] = delivery.UserId

	ctx = AppContext{
		ValueStore: store,
	}

	return ctx

}

func NewContext(r *http.Request) (ctx context.Context) {

	store := make(map[string]interface{})

	store[CorrelationIdHeader] = r.Header.Get(CorrelationIdHeader)
	store[AppIdHeader] = r.Header.Get(AppIdHeader)
	store[AuthorizedUserIDHeader] = r.Header.Get(AuthorizedUserIDHeader)

	ctx = AppContext{
		ValueStore: store,
	}

	return ctx

}

func NewContextFromValues(appID app.ApplicationID, correlationID string) (ctx context.Context) {

	store := make(map[string]interface{})

	store[CorrelationIdHeader] = correlationID
	store[AppIdHeader] = string(appID)

	ctx = AppContext{
		ValueStore: store,
	}

	return ctx

}

func NewContextFromValuesWithUser(appID app.ApplicationID, correlationID string, authUserID string) (ctx context.Context) {

	store := make(map[string]interface{})

	store[CorrelationIdHeader] = correlationID
	store[AppIdHeader] = string(appID)
	store[AuthorizedUserIDHeader] = authUserID

	ctx = AppContext{
		ValueStore: store,
	}

	return ctx

}

func CommitSqlTxFromContext(ctx context.Context) (err error) {

	ctx.Done()
	store := ctx.(AppContext).ValueStore

	val, ok := store[SqlTransactionKey]

	if !ok {
		// No Tx, do nothing
		return nil
	} else {
		tx := val.(*sql.Tx)

		err := tx.Commit()

		if err != nil {

			store[SqlTransactionKey] = nil

			return err
		}

		store[SqlTransactionKey] = nil
	}

	return nil
}

func RollbackSqlTxFromContext(ctx context.Context) (err error) {

	store := ctx.(AppContext).ValueStore

	val, ok := store[SqlTransactionKey]

	if !ok {
		// No Tx, do nothing
		return nil
	} else {
		tx := val.(*sql.Tx)

		err := tx.Rollback()

		if err != nil {

			store[SqlTransactionKey] = nil

			return err
		}

		store[SqlTransactionKey] = nil
	}

	return nil
}
