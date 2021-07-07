package tracing

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/HelloSundayMorning/apputils/app"
	"github.com/HelloSundayMorning/apputils/appctx"
	"github.com/HelloSundayMorning/apputils/appevent"
	"github.com/HelloSundayMorning/apputils/log"
	"github.com/aws/aws-xray-sdk-go/header"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
	"net/http"
)

type (
	TracingSegment func(ctx context.Context) (err error)

	WorkloadType string
)

const (
	correlationID = "correlationID"
	authUserID    = "authUserID"
	authUserRoles = "authUserRoles"

	workloadTypeAnnotationTitle = "WorkloadType"
	workloadGQLPath             = "GraphQLPath"
	workloadEventType           = "EventType"
	WorkloadTypeHTTPCall        = WorkloadType("HttpRequest")
	WorkloadTypeGraphQL         = WorkloadType("GraphQLRequest")
	WorkloadTypeGraphQLMutation = WorkloadType("GraphQLMutation")
	WorkloadTypeGraphQLQuery    = WorkloadType("GraphQLQuery")
	WorkloadTypeEventHandling   = WorkloadType("EventHandling")

	AWSXrayTraceId = "X-Amzn-Trace-Id"
)

// DefineTracingSegment
// Helper function allowing to define a closure for a specific code section that needs tracing
// The section will be identified inside AWS XRay as a subsegment
func DefineTracingSegment(ctx context.Context, segmentName string, funcTracingSegment TracingSegment) (err error) {

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

func AddTracingGraphQLInfo(ctx context.Context) {

	log.Printf(ctx, "AddTracingGraphQLInfo", "ctx in tracing: %+v", ctx)

	var pathStr string
	pathCtx := graphql.GetPathContext(ctx)

	if pathCtx != nil {
		pathStr = pathCtx.Path().String()
	}

	log.Printf(ctx, "AddTracingGraphQLInfo", "Request to GraphQL Path %s", pathStr)

	AddCustomTracingWorkloadType(ctx, WorkloadTypeGraphQL)

	_ = xray.AddAnnotation(ctx, workloadGQLPath, pathStr)

}

// GetParentSegmentTraceIDHeader
// Return a Xray header "Root=<trace>;Parent=<seg>;Sampled=<sample>" with the trace id and parent segment information from the context
// The context Segment info has to be added by a Xray Segment initialization called before,
// otherwise the function will return ""
func GetParentSegmentTraceIDHeader(ctx context.Context) (newHeader string) {

	seg := xray.GetSegment(ctx)

	if seg == nil {
		return ""
	}

	return seg.DownstreamHeader().String()

}

// BeginSegmentFromEventDelivery
// Return a XRay segment with the trace information and parent segment from the ampq delivery
// This enable event handling tracing within the same Xray TraceID and connect publisher and subscribers
// It expects the "X-Amzn-Trace-Id" header in format "Root=<trace>;Parent=<seg>;Sampled=<sample>"
// otherwise a new trace is started.
func BeginSegmentFromEventDelivery(ctx context.Context, appID app.ApplicationID, delivery amqp.Delivery) (context.Context, *xray.Segment) {

	var seg *xray.Segment

	if delivery.Headers[AWSXrayTraceId] != nil {
		log.Printf(ctx, "BeginSegmentFromEventDelivery", "Received tracing header from delivery: %s", delivery.Headers[AWSXrayTraceId].(string))

		xRayTraceHeader := header.FromString(delivery.Headers[AWSXrayTraceId].(string))

		// Here we create a dummy HTTP request to provide to XRay lib the required dependencies.
		// It's required since the library don't understand tracing from a parent segment
		// if it's not originating from a HTTP request.
		// This also guarantee that the sampling rule from the parent is propagated
		r, err := http.NewRequest("GET", "/", nil)

		if err != nil {
			// if an error occur while creating the dummy HTTP request, we just ignore it and
			// consider no request. As consequence it will prevent the sampling rule from the parent segment
			// to propagate to this segment
			r = nil
		}

		ctx, seg = xray.NewSegmentFromHeader(ctx, string(appID), r, xRayTraceHeader)

	} else {
		log.Printf(ctx, "BeginSegmentFromEventDelivery", "No tracing header from delivery found")

		ctx, seg = xray.BeginSegment(ctx, string(appID))
	}

	event, _ := appevent.NewAppEventFromJSON(delivery.Body)

	_ = xray.AddAnnotation(ctx, workloadEventType, event.EventType)

	AddCustomTracingWorkloadType(ctx, WorkloadTypeEventHandling)
	AddTracingAnnotationFromCtx(ctx)

	return ctx, seg
}
