package eventpubsub

import (
	"fmt"
	"github.com/HelloSundayMorning/apputils/appctx"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

type (

	PubSubTx interface {
		PublishToTopic(ctx context.Context, topic string, event []byte, contentType string) (err error)
		Commit() (err error)
		Rollback() (err error)
	}

	ChannelTx struct {
		publishChannel  *amqp.Channel
		registeredTopic map[string]bool
	}
)

func (chTx *ChannelTx) PublishToTopic(ctx context.Context, topic string, event []byte, contentType string) (err error) {

	appID := ctx.Value(appctx.AppIdHeader).(string)
	correlationID := ctx.Value(appctx.CorrelationIdHeader).(string)

	//valueUserID := ctx.Value(appctx.AuthorizedUserIDHeader)
	//userID := ""
	//
	//if valueUserID != nil {
	//	userID = valueUserID.(string)
	//}

	if !chTx.registeredTopic[topic] {
		return fmt.Errorf("app %s is not registered for topic %s", appID, topic)
	}

	msgID, err := uuid.NewV4()

	if err != nil {
		return fmt.Errorf("error getting uuid message ID, %s", err)
	}

	err = chTx.publishChannel.Publish(
		topic,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:   contentType,
			Body:          event,
			MessageId:     msgID.String(),
			DeliveryMode:  uint8(2),
			CorrelationId: correlationID,
			AppId:         string(appID),
			//UserId:        userID,
		})

	if err != nil {
		return fmt.Errorf("error publishing to channel, %s", err)
	}

	return nil
}

func (chTx *ChannelTx) Commit() (err error) {

	err = chTx.publishChannel.TxCommit()

	if err != nil {
		return err
	}

	return nil
}

func (chTx *ChannelTx) Rollback() (err error) {

	err = chTx.publishChannel.TxCommit()

	if err != nil {
		return err
	}

	return nil
}
