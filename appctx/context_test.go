package appctx

import (
	"net/http"
	"strings"
	"testing"

	"github.com/HelloSundayMorning/apputils/app"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"

	"github.com/HelloSundayMorning/hsmevents/roles" // now hsmenvet in go mod
)

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
		Headers: amqp.Table{
			AuthorizedUserIDHeader :"userID",
			AuthorizedUserRolesHeader: "Admin,Test",
		},
	}

	ctx := NewContextFromDelivery("AppID", delivery)

	sessionID := ctx.Value(CorrelationIdHeader).(string)
	appID := ctx.Value(AppIdHeader).(string)
	fromAppID := ctx.Value(FromAppIdHeader).(string)
	userID := ctx.Value(AuthorizedUserIDHeader).(string)
	roles := ctx.Value(AuthorizedUserRolesHeader).(string)

	assert.Equal(t, "CorrelationID", sessionID)
	assert.Equal(t, "AppID", appID)
	assert.Equal(t, "FromAppID", fromAppID)
	assert.Equal(t, "userID", userID)
	assert.Equal(t, "Admin,Test", roles)
}

func TestNewContextFromValue(t *testing.T) {

	ctx := NewContextFromValues(app.ApplicationID("AppID"), "CorrelationID")

	sessionID := ctx.Value(CorrelationIdHeader).(string)
	appID := ctx.Value(AppIdHeader).(string)

	assert.Equal(t, "CorrelationID", sessionID)
	assert.Equal(t, "AppID", appID)

}

func TestHasAllowRoles_OnlyAdminAndTriageAllowed(t *testing.T) {

	tests := []struct {
		Roles    []string
		Expected bool
		Msg      string
	}{
		{
			Roles:    []string{},
			Expected: false,
			Msg:      "disallow: empty roles",
		}, {
			Roles:    []string{roles.UserRoleAdmin, roles.UserRoleTriageUser},
			Expected: true,
			Msg:      "fully allow: admin and triage user",
		}, {
			Roles:    []string{roles.UserRoleFeedMonitor},
			Expected: false,
			Msg:      "fully disallowed: feed monitor",
		}, {
			Roles:    []string{roles.UserRoleCommunityManager},
			Expected: false,
			Msg:      "fully disallowed: community manager",
		}, {
			Roles:    []string{roles.UserRoleTriageUser, roles.UserRoleCommunityManager},
			Expected: true,
			Msg:      "partiall allowed, partial disallowed : community manager + triage user should be allowed",
		},
	}

	for _, test := range tests {

		// Test triage app permission
		ctx := NewContextFromValuesWithUserRoles(
			"testApp", "", "",
			strings.Join([]string{roles.UserRoleTriageUser, roles.UserRoleAdmin}, ","))

		actual := HasAllowRoles(ctx, test.Roles)

		assert.Equal(t, test.Expected, actual, test.Msg)
	}
}
