package log

import (
	"encoding/json"
	"fmt"
	"github.com/HelloSundayMorning/apputils/appctx"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type MyJSONFormatter struct {
}

func (f *MyJSONFormatter) Format(entry *log.Entry) ([]byte, error) {
	// Note this doesn't include Time, Level and Message which are available on
	// the Entry. Consult `godoc` on information about those fields or read the
	// source of the official loggers.
	serialized, err := json.Marshal(entry.Data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

func init() {
	//log.SetFormatter(new(MyJSONFormatter))
}


func Errorf(ctx context.Context, component string, format string, args ...interface{}) {

	log.WithFields(fields(ctx, component)).Errorf(format, args...)

}

func ErrorfNoContext(appID, component string, format string, args ...interface{}) {

	log.WithFields(fieldsNoContext(appID, component)).Errorf(format, args...)

}

func Printf(ctx context.Context, component string, format string, args ...interface{}) {

	log.WithFields(fields(ctx, component)).Printf(format, args...)

}

func PrintfNoContext(appID, component string, format string, args ...interface{}) {

	log.WithFields(fieldsNoContext(appID, component)).Printf(format, args...)

}

func Fatalf(ctx context.Context, component string, format string, args ...interface{}) {

	log.WithFields(fields(ctx, component)).Fatalf(format, args...)

}

func FatalfNoContext(appID, component string, format string, args ...interface{}) {

	log.WithFields(fieldsNoContext(appID, component)).Fatalf(format, args...)

}

func fields(ctx context.Context, component string) log.Fields {

	appID := ""
	correlationID := ""

	if ctx.Value(appctx.AppIdHeader) != nil {
		appID = ctx.Value(appctx.AppIdHeader).(string)
	}

	if ctx.Value(appctx.CorrelationIdHeader) != nil {
		correlationID = ctx.Value(appctx.CorrelationIdHeader).(string)
	}

	return log.Fields{
		"appId":         appID,
		"correlationId": correlationID,
		"component":     component,
	}

}

func fieldsNoContext(appID, component string) log.Fields {

	return log.Fields{
		"appId":         appID,
		"component":     component,
	}

}
