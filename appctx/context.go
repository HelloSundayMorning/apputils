package appctx

import (
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
	"net/http"
	"reflect"
	"time"
)

const(
	CorrelationIdHeader = "x-correlation-id"
	AppIdHeader         = "x-app-id"
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

func NewContextFromDelivery(delivery amqp.Delivery) (ctx context.Context) {

	store := make(map[string]interface{})

	store[CorrelationIdHeader] = delivery.CorrelationId
	store[AppIdHeader] = delivery.AppId

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

func NewContextFromValues(app, correlationID string) (ctx context.Context) {

	store := make(map[string]interface{})

	store[CorrelationIdHeader] = app
	store[AppIdHeader] = correlationID

	ctx = AppContext{
		ValueStore: store,
	}

	return ctx

}


