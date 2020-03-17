package eventpubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/HelloSundayMorning/apputils/app"
	"github.com/HelloSundayMorning/apputils/appctx"
	"github.com/HelloSundayMorning/apputils/log"
	uuid "github.com/satori/go.uuid"

	// Should all use wabbit
	"github.com/PeriscopeData/wabbit"
	"github.com/PeriscopeData/wabbit/amqp"
	srv "github.com/PeriscopeData/wabbit/amqptest/server"
)

type (
	MockMQ struct {
		AppID                app.ApplicationID
		MqConnection         *wabbit.Conn // interface
		registeredTopic      map[string]bool
		publishChannel       *wabbit.Channel // interface
		subscriptionChannels map[string]chan bool
	}
)

const (
	CorrelationIdHeader    = "x-correlation-id"
	AppIdHeader            = "x-app-id"
	FromAppIdHeader        = "x-from-app-id"
	AuthorizedUserIDHeader = "x-authorized-user-id"
)

// mock ctx with userId
func MockContextFromDelivery(appID app.ApplicationID, delivery wabbit.Delivery) (ctx context.Context) {
	userID := ""

	valUserID := delivery.Headers()[AuthorizedUserIDHeader]
	if valUserID != nil {
		userID = valUserID.(string)
	}

	ctx = context.Background()
	ctx = context.WithValue(ctx, CorrelationIdHeader, delivery.CorrelationId)
	ctx = context.WithValue(ctx, AppIdHeader, string(appID))
	ctx = context.WithValue(ctx, FromAppIdHeader, appID)
	ctx = context.WithValue(ctx, AuthorizedUserIDHeader, userID)

	return ctx
}

func NewWabbitMq(appID app.ApplicationID, user, pw, host string) (mockMQ *MockMQ, err error) {

	url := fmt.Sprintf("amqp://%s:%s@%s", user, pw, host)

	mqConnection, err := amqp.Dial(url)

	if err != nil {
		return mockMQ, fmt.Errorf("fail to dial WabbitMQ %s, %s", url, err)
	}

	mockMQ = &MockMQ{
		AppID:                appID,
		MqConnection:         &mqConnection,
		registeredTopic:      make(map[string]bool),
		subscriptionChannels: make(map[string]chan bool),
	}

	mockMQ.watchConnection()

	return mockMQ, nil
}

// this method similar to rabbitmq, use by NewWabbitMq
func (mockMQ *MockMQ) watchConnection() {

	receiver := make(chan wabbit.Error)

	conn := *mockMQ.MqConnection

	// pass value not pointer
	receiver = conn.NotifyClose(receiver)

	go func() {
		for {
			select {
			case rErr := <-receiver:
				if rErr == nil {
					log.PrintfNoContext(mockMQ.AppID, component, "WabbitMQ Connection closed")
					return
				} else {
					log.FatalfNoContext(mockMQ.AppID, component, "WabbitMQ Connection error, terminating app..., %s", rErr)
					return
				}
			}
		}
	}()

}

func (mockMQ *MockMQ) CleanUp() error {

	for _, subChannel := range mockMQ.subscriptionChannels {
		close(subChannel)
	}

	mockMQ.subscriptionChannels = make(map[string]chan bool, 0)

	mockMQ.registeredTopic = make(map[string]bool)

	if mockMQ.publishChannel != nil {

		channel := *mockMQ.publishChannel

		err := channel.Close()

		if err != nil {
			return fmt.Errorf(
				"error closing public channel while cleaning up Wabbitmq connection, %s",
				err)
		}
	}

	return nil
}

func (mockMQ *MockMQ) RegisterTopic(topic string) (err error) {
	// wabbit conn
	channel, err := (*mockMQ.MqConnection).Channel()
	if err != nil {
		return err
	}

	defer func() {
		err := channel.Close()
		if err != nil {
			log.ErrorfNoContext(mockMQ.AppID, component, "Error closing channel while registering topic, %s", err)
		}
	}()

	// Topic exchange
	err = channel.ExchangeDeclare(
		topic,
		"fanout",
		wabbit.Option{
			"durable":    true,
			"autoDelete": false,
			"internal":   false,
			"noWait":     false,
			"args":       nil,
		},
	)

	if err != nil {
		return err
	}

	mockMQ.registeredTopic[topic] = true

	log.PrintfNoContext(mockMQ.AppID, component, "Registered topic %s for app %s", topic, mockMQ.AppID)
	return nil
}

func (mockMQ *MockMQ) InitializeQueue(topic string) (err error) {

	for attempts := 1; attempts < 4; attempts++ {
		// declare queue
		err = mockMQ.declareQueue(topic)
		if err == nil {
			break
		}

		log.PrintfNoContext(mockMQ.AppID, component, "Failed to initialize queue for topic %s. Waiting (30 sec) for next attempt. Total Attempts = %d", topic, attempts)

		time.Sleep(time.Second * 30)
	}

	if err != nil {
		log.PrintfNoContext(mockMQ.AppID, component, "Failed to initialize queue for topic %s. %s", topic, err)
		return err
	}

	return nil
}

func (mockMQ *MockMQ) declareQueue(topic string) (err error) {

	channel, err := (*mockMQ.MqConnection).Channel()
	if err != nil {
		return err
	}

	defer func() {
		err := channel.Close()
		if err != nil {
			log.ErrorfNoContext(mockMQ.AppID, component, "Error closing channel while initializing queue, %s", err)
		}
	}()

	appQueueName := formQueueName(mockMQ.AppID, topic)

	deadLetterName := formDeadLetterName(mockMQ.AppID, topic)

	// topic exchange
	err = channel.ExchangeDeclare(
		deadLetterName,
		"fanout",
		wabbit.Option{
			"durable":    true,
			"autoDelete": false,
			"internal":   false,
			"noWait":     false,
			"args":       nil,
		},
	)

	if err != nil {
		return err
	}

	_, err = MockNewFanOutQueue(&channel, deadLetterName, deadLetterName, "")
	if err != nil {
		return fmt.Errorf("error creating dead letter queue: %s", err)
	}

	_, err = MockNewFanOutQueue(&channel, topic, appQueueName, deadLetterName)
	if err != nil {
		return fmt.Errorf("error creating queue: %s", err)
	}

	return nil
}

// Mock lib: wabbit don't implement Tx(), TxCommit(), TxRollback() for Channel struct
// ChannelTx is our high level wrapper of Channel, also need mock
// So now, leave it as we not use often
// func (mockMQ *MockMQ) PublishWithTx(txFunc PublishTxHandler) (err error) {
//
//     ch, err := (*mockMQ.MqConnection).Channel()
//     if err != nil {
//         return err
//     }
//
//     defer func() {
//         _ = ch.Close()
//     }()
//
//     err = ch.Tx()
//
//     if err != nil {
//         return err
//     }
//
//     err = txFunc(&ChannelTx{
//         publishChannel: &ch.,
//         registeredTopic: mockMQ.registeredTopic,
//     })
//
//     if err != nil {
//         _ = ch.TxRollback()
//         return err
//     }
//
//     err = ch.TxCommit()
//
//     if err != nil {
//         _ = ch.TxRollback()
//         return err
//     }
//
//     return nil
// }

func (mockMQ *MockMQ) PublishToTopic(ctx context.Context, topic string, event []byte, contentType string) (err error) {

	appID := ctx.Value(appctx.AppIdHeader).(string)
	correlationID := ctx.Value(appctx.CorrelationIdHeader).(string)

	valueUserID := ctx.Value(appctx.AuthorizedUserIDHeader)
	userID := ""

	if valueUserID != nil {
		userID = valueUserID.(string)
	}

	if !mockMQ.registeredTopic[topic] {
		return fmt.Errorf("app %s is not registered for topic %s", appID, topic)
	}

	if mockMQ.publishChannel == nil {
		err := mockMQ.newPublishChannel()
		if err != nil {
			return err
		}
	}

	msgID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("error getting uuid message ID, %s", err)
	}

	err = (*mockMQ.publishChannel).Publish(
		topic,
		"",
		event,
		wabbit.Option{
			"headers": map[string]interface{}{
				appctx.AuthorizedUserIDHeader: userID,
			},
			"contentType":   contentType,
			"correlationId": correlationID,
			"deliveryMode":  uint8(2),
			"messageId":     msgID.String(),
		},
	)

	if err != nil {
		return err
	}

	return nil
}

// assign channel in connection to publishChannel
func (mockMQ *MockMQ) newPublishChannel() (err error) {

	channel, err := (*mockMQ.MqConnection).Channel()
	if err != nil {
		return err
	}

	mockMQ.publishChannel = &channel

	log.PrintfNoContext(mockMQ.AppID, component, "New channel set to publish channel.")
	return nil
}

// mock RabbitMQ object
func (mockMQ *MockMQ) SubscribeToTopic(topic string, processFunc ProcessEvent) (err error) {
	return mockMQ.SubscribeToTopicWithMaxMsg(topic, processFunc, 0)
}

func (mockMQ *MockMQ) SubscribeToTopicWithMaxMsg(topic string, processFunc ProcessEvent, maxMessages int) (err error) {

	appQueueName := formQueueName(mockMQ.AppID, topic)

	// convert interface to conn struct, which have more methods
	mockConn := (*mockMQ.MqConnection).(*amqp.Conn)

	// wabbit channel
	channel, err := mockConn.Channel()
	if err != nil {
		return err
	}

	if maxMessages != 0 {
		err = channel.Qos(maxMessages, 0, false)
		if err != nil {
			return fmt.Errorf("invalid Qos setup, %s", err)
		}
	}

	deliveries, err := channel.Consume(appQueueName, "",
		wabbit.Option{
			"autoAck":   false,
			"exclusive": false,
			"noLocal":   false,
			"noWait":    false,
		},
	)
	if err != nil {
		return err
	}

	mockMQ.subscriptionChannels[topic] = make(chan bool)

	go func() {
		for {
			select {
			case delivery := <-deliveries:
				// Internal function below
				mockMQ.handleDelivery(delivery, processFunc)
			case <-mockMQ.subscriptionChannels[topic]:
				// Close
				err := channel.Close()
				if err != nil {
					log.ErrorfNoContext(mockMQ.AppID, component, "Error closing channel while ending subscription, %s", err)
				}
				return
			}
		}
	}()

	log.PrintfNoContext(mockMQ.AppID, component, "App %s Subscribed to topic %s", mockMQ.AppID, topic)
	return nil
}

func (mockMQ *MockMQ) UnSubscribe(topic string) {
	subChan := mockMQ.subscriptionChannels[topic]
	if subChan != nil {
		close(subChan)
	}
}

// send to DLQ if fail 2nd , amqp.Delivery is a struct
func (mockMQ *MockMQ) handleDelivery(delivery wabbit.Delivery, processFunc ProcessEvent) {

	if delivery.CorrelationId() == "" {
		id, _ := uuid.NewV4()
		d := delivery.(*srv.Delivery)
		d.SetCorrelationID(id.String())
	}

	ctx := MockContextFromDelivery(mockMQ.AppID, delivery)

	err := processFunc(ctx, delivery.Body(), "application/json")
	if err != nil {

		log.Errorf(ctx, component, "Error handling delivery, %s", err)

		if delivery.Redelivered() {
			log.Printf(ctx, component, "2nd attempt failure. Dead-letter delivery, %s", err)
			err = delivery.Nack(false, false)
		} else {
			log.Printf(ctx, component, "1st attempt failure. Re-queue delivery, %s", err)
			err = delivery.Nack(false, true)
		}

		if err != nil {
			log.Errorf(ctx, component, "Error while Nack delivery, %s", err)
		}
		return
	}

	err = delivery.Ack(false)
	if err != nil {
		log.Errorf(ctx, component, "Error while Ack delivery, %s", err)
	}
}

func MockNewFanOutQueue(channelPointer *wabbit.Channel, exchangeName, queueName string, deadLetterExchange string) (queue wabbit.Queue, err error) {

	channel := *channelPointer

	// topic exchange
	err = channel.ExchangeDeclarePassive(
		exchangeName,
		"fanout",
		wabbit.Option{
			"durable":    true,
			"autoDelete": false,
			"internal":   false,
			"noWait":     false,
			"args":       nil,
		},
	)

	if err != nil {
		return queue, fmt.Errorf("could not find exchange %s to bind queue %s, %s", exchangeName, queueName, err)
	}

	args := make(map[string]interface{})

	if deadLetterExchange != "" {
		// dead letter queue name where a failed msg is sent
		args["x-dead-letter-exchange"] = deadLetterExchange
	} else {
		// set dead letter queue to mode lazy to store messages in disk not memory
		args["x-queue-mode"] = "lazy"
	}

	// topic durable queue
	queue, err = channel.QueueDeclare(
		queueName,
		wabbit.Option{
			"durable":    true,
			"autoDelete": false,
			"internal":   false,
			"noWait":     false,
			"args":       args,
		},
	)

	if err != nil {
		return queue, err
	}

	err = channel.QueueBind(
		queue.Name(),
		"",
		exchangeName,
		wabbit.Option{
			"noWait": false,
			"args":   nil,
		},
	)

	if err != nil {
		return queue, err
	}

	return queue, nil
}
