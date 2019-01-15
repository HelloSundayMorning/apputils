package eventpubsub

import (
	"context"
	"fmt"
	"github.com/HelloSundayMorning/apputils/appctx"
	"github.com/HelloSundayMorning/apputils/log"
	"github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type (
	RabbitMq struct {
		MqConnection         *amqp.Connection
		registeredTopic      map[string]bool
		publishChannel       *amqp.Channel
		subscriptionChannels map[string]chan bool
	}
)

func NewRabbitMq(user, pw, host string) (handler *RabbitMq, err error) {

	url := fmt.Sprintf("amqp://%s:%s@%s", user, pw, host)

	mqConnection, err := amqp.Dial(url)

	if err != nil {
		return handler, fmt.Errorf("fail to dial RabbitMQ %s, %s", url, err)
	}

	handler = &RabbitMq{
		MqConnection:         mqConnection,
		registeredTopic:      make(map[string]bool),
		subscriptionChannels: make(map[string]chan bool),
	}

	return handler, nil

}

func (rabbit *RabbitMq) CleanUp() {

	for _, subChannel := range rabbit.subscriptionChannels {
		close(subChannel)
	}

	rabbit.subscriptionChannels = make(map[string]chan bool, 0)

	if rabbit.publishChannel != nil {
		rabbit.publishChannel.Close()
	}

	rabbit.registeredTopic = make(map[string]bool)

}

// RegisterTopic Should be called in the initialization to create an exchange
// If the exchange exists it's ignored
func (rabbit *RabbitMq) RegisterTopic(appID, topic string) (err error) {

	channel, err := rabbit.MqConnection.Channel()

	defer channel.Close()

	if err != nil {
		return err
	}

	// topic exchange
	err = channel.ExchangeDeclare(
		topic,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	rabbit.registeredTopic[topic] = true

	return nil
}

// InitializeQueue should be called at the application start for each topic the app
// will subscribe to.
// It will create the topic queue for the subscription and the dead
// letter queue, with the subscription for the dead letter handler
//
// Attempt to subscribe to a non existent topic will return a error
//
// - appID : unique name for the application that will subscribe to a topic
// - topic : topic name
func (rabbit *RabbitMq) InitializeQueue(appID, topic string) (err error) {

	channel, err := rabbit.MqConnection.Channel()

	defer channel.Close()

	if err != nil {
		return err
	}

	appQueueName := formQueueName(appID, topic)
	deadLetterName := formDeadLetterName(appID, topic)

	// topic exchange
	err = channel.ExchangeDeclare(
		deadLetterName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	_, err = newFanOutQueue(channel, deadLetterName, deadLetterName, "")

	if err != nil {
		return fmt.Errorf("error creating dead letter queue: %s", err)
	}

	_, err = newFanOutQueue(channel, topic, appQueueName, deadLetterName)

	if err != nil {
		return fmt.Errorf("error creating queue: %s", err)
	}

	return nil
}

func (rabbit *RabbitMq) Publish(ctx context.Context, topic string, event []byte, contentType string) (err error) {

	appID := ctx.Value(appctx.AppIdHeader).(string)
	correlationID := ctx.Value(appctx.CorrelationIdHeader).(string)

	if !rabbit.registeredTopic[topic] {
		return fmt.Errorf("app %s is not registered for topic %s", appID, topic)
	}

	if rabbit.publishChannel == nil {

		channel, err := rabbit.MqConnection.Channel()

		if err != nil {
			return err
		}

		rabbit.publishChannel = channel

	}

	err = rabbit.publishChannel.ExchangeDeclarePassive(
		topic,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	msgID, err := uuid.NewV4()

	if err != nil {
		return fmt.Errorf("error getting uuid message ID, %s", err)
	}

	err = rabbit.publishChannel.Publish(
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
			AppId:         appID,
		})

	if err != nil {
		return err
	}

	return nil
}

func (rabbit *RabbitMq) Subscribe(appID, topic string, processFunc ProcessEvent) (err error) {

	return rabbit.SubscribeWithMaxMsg(appID, topic, processFunc, 0)
}

func (rabbit *RabbitMq) SubscribeWithMaxMsg(appID, topic string, processFunc ProcessEvent, maxMessages int) (err error) {

	appQueueName := formQueueName(appID, topic)

	channel, err := rabbit.MqConnection.Channel()

	if err != nil {
		return err
	}

	if maxMessages != 0 {
		err = channel.Qos(maxMessages, 0, false)

		if err != nil {
			return fmt.Errorf("invalid Qos setup, %s", err)
		}
	}

	deliveries, err := channel.Consume(
		appQueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	rabbit.subscriptionChannels[topic] = make(chan bool)

	go func() {
		for {
			select {
			case delivery := <-deliveries:

				handleDelivery(appID, delivery, processFunc)

			case <-rabbit.subscriptionChannels[topic]:

				channel.Close()

				return

			}
		}
	}()

	return nil
}

func (rabbit *RabbitMq) UnSubscribe(topic string) {

	subChan := rabbit.subscriptionChannels[topic]

	if subChan != nil {
		close(subChan)
	}

}

// handleDelivery
// Call process function. If it fails requeue the first time.
// the second fail will send it to dead letter
func handleDelivery(appID string, delivery amqp.Delivery, processFunc ProcessEvent) {

	if delivery.CorrelationId == "" {
		id, _ := uuid.NewV4()
		delivery.CorrelationId = id.String()
	}

	ctx := appctx.NewContextFromDelivery(appID, delivery)

	err := processFunc(ctx, delivery.Body, delivery.ContentType)

	if err != nil {

		log.Errorf(ctx, "eventpubsub_rabbitmq", "error handling delivery, %s", err)

		if delivery.Redelivered {
			delivery.Nack(false, false)
		} else {
			delivery.Nack(false, true)
		}

		return
	}

	delivery.Ack(false)
}

func formQueueName(appID, topic string) string {
	return fmt.Sprintf("%s->%s", appID, topic)

}

func formDeadLetterName(appID, topic string) string {
	return fmt.Sprintf("%s->%s.deadletter", appID, topic)
}

// newFanOutQueue
// creates:
// - a durable queue for a exchange
//
// If exchange is not found retry 3 times to find it with a interval of a 30 sec.
//
func newFanOutQueue(channel *amqp.Channel, exchangeName, queueName string, deadLetterExchange string) (queue amqp.Queue, err error) {

	// topic exchange
	err = channel.ExchangeDeclarePassive(
		exchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return queue, fmt.Errorf("could not find exchange %s to bind queue %s, %s", exchangeName, queueName, err)
	}

	args := make(map[string]interface{})

	if deadLetterExchange != "" {
		args["x-dead-letter-exchange"] = deadLetterExchange //dead letter queue name where a failed msg is sent
	} else {
		args["x-queue-mode"] = "lazy" // set dead letter queue to mode lazy to store messages in disk not memory
	}

	// topic durable queue
	queue, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		args,
	)

	if err != nil {
		return queue, err
	}

	err = channel.QueueBind(
		queue.Name,
		"",
		exchangeName,
		false,
		nil,
	)

	if err != nil {
		return queue, err
	}

	return queue, nil
}
