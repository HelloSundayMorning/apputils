package log

import (
	"encoding/json"
	"fmt"
	"github.com/HelloSundayMorning/apputils/app"
	"github.com/HelloSundayMorning/apputils/appctx"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type MyJSONFormatter struct {
}

func (f *MyJSONFormatter) Format(entry *log.Entry) ([]byte, error) {

	serialized, err := json.Marshal(entry.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

func init() {
	//log.SetFormatter(new(MyJSONFormatter))
}

func Errorf(ctx context.Context, component string, format string, args ...interface{}) {

	log.WithFields(fields(ctx, component)).Errorf(format, args...)

}

func ErrorfNoContext(appID app.ApplicationID, component string, format string, args ...interface{}) {

	log.WithFields(fieldsNoContext(appID, component)).Errorf(format, args...)

}

func Printf(ctx context.Context, component string, format string, args ...interface{}) {

	log.WithFields(fields(ctx, component)).Printf(format, args...)

}

func PrintfNoContext(appID app.ApplicationID, component string, format string, args ...interface{}) {

	log.WithFields(fieldsNoContext(appID, component)).Printf(format, args...)

}

func Fatalf(ctx context.Context, component string, format string, args ...interface{}) {

	log.WithFields(fields(ctx, component)).Fatalf(format, args...)

}

func FatalfNoContext(appID app.ApplicationID, component string, format string, args ...interface{}) {

	log.WithFields(fieldsNoContext(appID, component)).Fatalf(format, args...)

}

func fields(ctx context.Context, component string) log.Fields {

	appID := ""
	correlationID := ""
	UserID := ""
	Roles := ""

	if ctx.Value(appctx.AppIdHeader) != nil {
		appID = ctx.Value(appctx.AppIdHeader).(string)
	}

	if ctx.Value(appctx.CorrelationIdHeader) != nil {
		correlationID = ctx.Value(appctx.CorrelationIdHeader).(string)
	}

	if ctx.Value(appctx.AuthorizedUserIDHeader) != nil {
		UserID = ctx.Value(appctx.AuthorizedUserIDHeader).(string)
	}

	if ctx.Value(appctx.AuthorizedUserRolesHeader) != nil {
		Roles = ctx.Value(appctx.AuthorizedUserRolesHeader).(string)
	}

	return log.Fields{
		"appId":         appID,
		"correlationId": correlationID,
		"component":     component,
		"authUserID":    UserID,
		"authUserRoles": Roles,
	}

}

func fieldsNoContext(appID app.ApplicationID, component string) log.Fields {

	return log.Fields{
		"appId":     appID,
		"component": component,
	}

}
