package tracing

import (
	"github.com/HelloSundayMorning/apputils/appctx"
	"github.com/aws/aws-xray-sdk-go/xray"
	"golang.org/x/net/context"
)

type (
	TracingSegment func(ctx context.Context)
)

const (
	correlationID = "correlationID"
	authUserID    = "authUserID"
	authUserRoles = "authUserRoles"
)

func DefineTracingSegment(ctx context.Context, segmentName string, funcTracingSegment TracingSegment) {

	_, subSeg := xray.BeginSubsegment(ctx, segmentName)

	funcTracingSegment(ctx)

	subSeg.Close(nil)

}

func AddTracingAnnotationFromCtx(ctx context.Context) {

	if ctx.Value(appctx.CorrelationIdHeader) != nil {
		_ = xray.AddAnnotation(ctx, correlationID, ctx.Value(appctx.CorrelationIdHeader).(string))
	}

	if ctx.Value(appctx.AuthorizedUserIDHeader) != nil {
		_ = xray.AddAnnotation(ctx, authUserID, ctx.Value(appctx.AuthorizedUserIDHeader).(string))
	}

	if ctx.Value(appctx.AuthorizedUserRolesHeader) != nil {
		_ = xray.AddMetadata(ctx, authUserRoles, ctx.Value(appctx.AuthorizedUserRolesHeader).(string))
	}
}
