package tracing

import (
	"fmt"
	"github.com/aws/aws-xray-sdk-go/header"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
)

func TestGetParentSegmentTraceIDHeader(t *testing.T) {

	ctx := context.Background()
	seg := &xray.Segment{
		TraceID: "TraceID",
		ID:      "SegID",
		Name:    "Test",
		IncomingHeader: &header.Header{
			TraceID:          "TraceID",
			ParentID:         "",
			SamplingDecision: header.Sampled,
			AdditionalData:   nil,
		},
		Sampled: true,
	}

	seg.ParentSegment = seg // set itself as root

	ctx = context.WithValue(ctx, xray.ContextKey, seg)

	TraceHeader := GetParentSegmentTraceIDHeader(ctx)

	assert.Equal(t, "Root=TraceID;Parent=SegID;Sampled=1", TraceHeader)

}

func TestBeginSegmentFromEventDelivery(t *testing.T) {

	traceID := xray.NewTraceID()
	segID := xray.NewSegmentID()

	ctx := context.Background()

	delivery := amqp.Delivery{
		AppId:         "FromAppID",
		CorrelationId: "CorrelationID",
		Headers: amqp.Table{
			AWSXrayTraceId: fmt.Sprintf("Root=%s;Parent=%s;Sampled=1", traceID, segID),
		},
	}

	ctx, seg := BeginSegmentFromEventDelivery(ctx, "appID", delivery)

	assert.Equal(t, traceID, seg.TraceID)
	assert.Equal(t, segID, seg.ParentID)
	assert.Equal(t, true, seg.Sampled)
	assert.NotEqual(t, segID, seg.ID)
}



