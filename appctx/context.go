package appctx

import (
	"github.com/HelloSundayMorning/apputils/app"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
	"net/http"
	"reflect"
	"time"
)

const(
	CorrelationIdHeader = "x-correlation-id"
	AppIdHeader         = "x-app-id"
	FromAppIdHeader     = "x-from-app-id"
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
	return nil
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

func NewContextFromDelivery(appId app.ApplicationID, delivery amqp.Delivery) (ctx context.Context) {

	store := make(map[string]interface{})

	store[CorrelationIdHeader] = delivery.CorrelationId
	store[AppIdHeader] = appId
	store[FromAppIdHeader] = delivery.AppId

	ctx = AppContext{
		ValueStore: store,
	}

	return ctx

}

func NewContext(r *http.Request) (ctx context.Context) {

	store := make(map[string]interface{})

	store[CorrelationIdHeader] = r.Header.Get(CorrelationIdHeader)
	store[AppIdHeader] = r.Header.Get(AppIdHeader)

	ctx = AppContext{
		ValueStore: store,
	}

	return ctx

}

func NewContextFromValues(appID app.ApplicationID, correlationID string) (ctx context.Context) {

	store := make(map[string]interface{})

	store[CorrelationIdHeader] = correlationID
	store[AppIdHeader] = appID

	ctx = AppContext{
		ValueStore: store,
	}

	return ctx

}


