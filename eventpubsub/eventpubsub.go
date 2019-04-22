package eventpubsub

import (
	"golang.org/x/net/context"
)

type (

	ProcessEvent func(ctx context.Context, event []byte, contentType string) error

	EventPubSub interface {
		RegisterTopic(appID, topic string) (err error)
		InitializeQueue(appID, topic string) (err error)
		Publish(ctx context.Context, topic string, event []byte, contentType string) (err error)
		Subscribe(appID, topic string, processFunc ProcessEvent) (err error)
		SubscribeWithMaxMsg(appID, topic string, processFunc ProcessEvent, maxMessages int) (err error)
		UnSubscribe(topic string)
		CleanUp() (err error)
	}

)


