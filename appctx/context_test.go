package appctx

import (
	"github.com/HelloSundayMorning/apputils/app"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestAppContext_Value(t *testing.T) {

	ctx := AppContext{
		ValueStore: make(map[string]interface{}),
	}

	ctx.ValueStore["key"] = "value"

	result := ctx.Value(1)

	assert.Nil(t, result)

	result = ctx.Value("key")

	assert.Equal(t, "value", result.(string))
}

func TestNewContext(t *testing.T) {

	r, _ := http.NewRequest("GET", "/", nil)

	r.Header.Set(CorrelationIdHeader,"CorrelationID")
	r.Header.Set(AppIdHeader,"AppID")

	ctx := NewContext(r)

	sessionID := ctx.Value(CorrelationIdHeader).(string)
	appID := ctx.Value(AppIdHeader).(string)

	assert.Equal(t, "CorrelationID", sessionID)
	assert.Equal(t, "AppID", appID)
}

func TestNewContextFromDelivery(t *testing.T) {

	delivery := amqp.Delivery{
		AppId: "FromAppID",
		CorrelationId: "CorrelationID",
	}

	ctx := NewContextFromDelivery("AppID", delivery)

	sessionID := ctx.Value(CorrelationIdHeader).(string)
	appID := ctx.Value(AppIdHeader).(string)
	fromAppID := ctx.Value(FromAppIdHeader).(string)

	assert.Equal(t, "CorrelationID", sessionID)
	assert.Equal(t, "AppID", appID)
	assert.Equal(t, "FromAppID", fromAppID)
}

func TestNewContextFromValue(t *testing.T) {

	ctx := NewContextFromValues(app.ApplicationID("AppID"), "CorrelationID")

	sessionID := ctx.Value(CorrelationIdHeader).(string)
	appID := ctx.Value(AppIdHeader).(string)

	assert.Equal(t, "CorrelationID", sessionID)
	assert.Equal(t, "AppID", appID)

}

