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
	AuthorizedUserRolesHeader     = "x-authorized-user-roles"


)

func NewContextFromDelivery(appID app.ApplicationID, delivery amqp.Delivery) (ctx context.Context) {

	valUserID := delivery.Headers[AuthorizedUserIDHeader]
	userID := ""

	if valUserID != nil {
		userID = valUserID.(string)
	}

	valUserRoles := delivery.Headers[AuthorizedUserRolesHeader]
	userRoles := ""

	if valUserRoles != nil {
		userRoles = valUserRoles.(string)
	}

	ctx = context.Background()

	ctx = context.WithValue(ctx, CorrelationIdHeader, delivery.CorrelationId)
	ctx = context.WithValue(ctx, AppIdHeader, string(appID))
	ctx = context.WithValue(ctx, FromAppIdHeader, delivery.AppId)
	ctx = context.WithValue(ctx, AuthorizedUserIDHeader, userID)
	ctx = context.WithValue(ctx, AuthorizedUserRolesHeader, userRoles)

	return ctx

}

func NewContext(r *http.Request) (ctx context.Context) {

	ctx = context.Background()

	ctx = context.WithValue(ctx, CorrelationIdHeader, r.Header.Get(CorrelationIdHeader))
	ctx = context.WithValue(ctx, AppIdHeader, r.Header.Get(AppIdHeader))
	ctx = context.WithValue(ctx, AuthorizedUserIDHeader, r.Header.Get(AuthorizedUserIDHeader))
	ctx = context.WithValue(ctx, AuthorizedUserRolesHeader, r.Header.Get(AuthorizedUserRolesHeader))

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

func NewContextFromValuesWithUserRoles(appID app.ApplicationID, correlationID string, authUserID, userRoles string) (ctx context.Context) {

	ctx = context.Background()

	ctx = context.WithValue(ctx, CorrelationIdHeader, correlationID)
	ctx = context.WithValue(ctx, AppIdHeader, string(appID))
	ctx = context.WithValue(ctx, AuthorizedUserIDHeader, authUserID)
	ctx = context.WithValue(ctx, AuthorizedUserRolesHeader, userRoles)

	return ctx

}

func GetAuthorizedUserID(ctx context.Context) (authorizedUserID string) {

	valueUserID := ctx.Value(AuthorizedUserIDHeader)

	if valueUserID != nil {
		authorizedUserID = valueUserID.(string)
	}

	return authorizedUserID

}

func GetAuthorizedUserRoles(ctx context.Context) (authorizedUserRoles string) {

	valueUserRoles := ctx.Value(AuthorizedUserRolesHeader)

	if valueUserRoles != nil {
		authorizedUserRoles = valueUserRoles.(string)
	}

	return authorizedUserRoles

}




