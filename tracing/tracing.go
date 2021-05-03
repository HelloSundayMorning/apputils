package tracing

import (
	"github.com/HelloSundayMorning/apputils/appctx"
	"github.com/aws/aws-xray-sdk-go/xray"
	"golang.org/x/net/context"
)

type (
	TracingSegment func(ctx context.Context) (err error)

	WorkloadType string
)

const (
	correlationID = "correlationID"
	authUserID    = "authUserID"
	authUserRoles = "authUserRoles"

	workloadTypeAnnotationTitle =  "WorkloadType"
	WorkloadTypeHTTPCall =  WorkloadType("HTTPCall")
	WorkloadTypeGraphQL =  WorkloadType("GraphQL")
	WorkloadTypeGraphQLMutation =  WorkloadType("GraphQLMutation")
	WorkloadTypeGraphQLQuery =  WorkloadType("GraphQLQuery")
	WorkloadTypeEventHandling =  WorkloadType("EventHandling")
)

// DefineTracingSegment
// Helper function allowing to define a closure for a specific code section that needs tracing
// The section will be identified inside AWS XRay as a subsegment
func DefineTracingSegment(ctx context.Context, segmentName string, funcTracingSegment TracingSegment) (err error){

	_, subSeg := xray.BeginSubsegment(ctx, segmentName)

	AddTracingAnnotationFromCtx(ctx)

	err = funcTracingSegment(ctx)

	subSeg.Close(err)

	return err

}

// AddTracingAnnotationFromCtx
// Add app context information to the AWS Xray segment as annotations and metadata
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

func AddCustomTracingWorkloadType(ctx context.Context, wt WorkloadType) {

	_ = xray.AddAnnotation(ctx, workloadTypeAnnotationTitle, string(wt))

}
