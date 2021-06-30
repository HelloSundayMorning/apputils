package eventpubsub

import (
	"golang.org/x/net/context"
)



type (

	ProcessEvent func(ctx context.Context, event []byte, contentType string) error
	PublishTxHandler func(tx PubSubTx) (err error)

	EventPubSub interface {
		RegisterTopic(topic string) (err error)
		InitializeQueue(topic string) (err error)
		PublishToTopic(ctx context.Context, topic string, event []byte, contentType string) (err error)
		SubscribeToTopic(topic string, processFunc ProcessEvent) (err error)
		SubscribeToTopicWithMaxMsg(topic string, processFunc ProcessEvent, maxMessages int) (err error)
		UnSubscribe(topic string)
		CleanUp() (err error)
		PublishWithTx(txFunc PublishTxHandler) (err error)
	}

)


