package appctx

import (
	"github.com/HelloSundayMorning/apputils/app"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
	"net/http"
)

const (
	CorrelationIdHeader           = "x-correlation-id"
	AppIdHeader                   = "x-app-id"
	FromAppIdHeader               = "x-from-app-id"
	AuthorizedUserIDHeader        = "x-authorized-user-id"


)

func NewContextFromDelivery(appID app.ApplicationID, delivery amqp.Delivery) (ctx context.Context) {

	valUserID := delivery.Headers[AuthorizedUserIDHeader]
	userID := ""

	if valUserID != nil {
		userID = valUserID.(string)
	}

	ctx = context.Background()

	ctx = context.WithValue(ctx, CorrelationIdHeader, delivery.CorrelationId)
	ctx = context.WithValue(ctx, AppIdHeader, string(appID))
	ctx = context.WithValue(ctx, FromAppIdHeader, delivery.AppId)
	ctx = context.WithValue(ctx, AuthorizedUserIDHeader, userID)

	return ctx

}

func NewContext(r *http.Request) (ctx context.Context) {

	ctx = context.Background()

	ctx = context.WithValue(ctx, CorrelationIdHeader, r.Header.Get(CorrelationIdHeader))
	ctx = context.WithValue(ctx, AppIdHeader, r.Header.Get(AppIdHeader))
	ctx = context.WithValue(ctx, AuthorizedUserIDHeader, r.Header.Get(AuthorizedUserIDHeader))

	return ctx

}

func NewContextFromValues(appID app.ApplicationID, correlationID string) (ctx context.Context) {

	ctx = context.Background()

	ctx = context.WithValue(ctx, CorrelationIdHeader, correlationID)
	ctx = context.WithValue(ctx, AppIdHeader, string(appID))

	return ctx

}

func NewContextFromValuesWithUser(appID app.ApplicationID, correlationID string, authUserID string) (ctx context.Context) {

	ctx = context.Background()

	ctx = context.WithValue(ctx, CorrelationIdHeader, correlationID)
	ctx = context.WithValue(ctx, AppIdHeader, string(appID))
	ctx = context.WithValue(ctx, AuthorizedUserIDHeader, authUserID)

	return ctx

}

func GetAuthorizedUserID(ctx context.Context) (authorizedUserID string) {

	valueUserID := ctx.Value(AuthorizedUserIDHeader)

	if valueUserID != nil {
		authorizedUserID = valueUserID.(string)
	}

	return authorizedUserID

}



