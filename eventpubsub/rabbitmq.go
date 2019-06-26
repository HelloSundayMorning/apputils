package eventpubsub

import (
	"fmt"
	"github.com/HelloSundayMorning/apputils/app"
	"github.com/HelloSundayMorning/apputils/appctx"
	"github.com/HelloSundayMorning/apputils/log"
	"github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
	"time"
)

type (
	RabbitMq struct {
		AppID                app.ApplicationID
		MqConnection         *amqp.Connection
		registeredTopic      map[string]bool
		publishChannel       *amqp.Channel
		subscriptionChannels map[string]chan bool
	}
)

const (
	component = "eventpubsub_rabbitmq"
)

func NewRabbitMq(appID app.ApplicationID, user, pw, host string) (rabbitMq *RabbitMq, err error) {

	url := fmt.Sprintf("amqp://%s:%s@%s", user, pw, host)

	mqConnection, err := amqp.Dial(url)

	if err != nil {
		return rabbitMq, fmt.Errorf("fail to dial RabbitMQ %s, %s", url, err)
	}

	rabbitMq = &RabbitMq{
		AppID:                appID,
		MqConnection:         mqConnection,
		registeredTopic:      make(map[string]bool),
		subscriptionChannels: make(map[string]chan bool),
	}

	rabbitMq.WatchConnection(appID, user, pw, host)

	return rabbitMq, nil

}

func (rabbit *RabbitMq) WatchConnection(appID app.ApplicationID, user, pw, host string) {

	receiver := make(chan *amqp.Error)

	receiver = rabbit.MqConnection.NotifyClose(receiver)

	go func() {
		for {
			select {
			case rErr := <-receiver:
				if rErr == nil {
					log.PrintfNoContext(rabbit.AppID, component, "RabbitMQ Connection closed")

					return
				} else {
					log.ErrorfNoContext(rabbit.AppID, component, "RabbitMQ Connection error, reconnecting..., %s", rErr)

					// preserve registered topics
					registeredTopics := make(map[string]bool)

					for k, v := range rabbit.registeredTopic {
						registeredTopics[k] = v
					}

					//TODO: preserve subscriptionChannels by re-initializing the queues and subscription channels

					_ = rabbit.CleanUp()

					rabbit, err := NewRabbitMq(appID, user, pw, host)

					rabbit.registeredTopic = registeredTopics

					if err != nil {
						log.FatalfNoContext(rabbit.AppID, component, "Error reconnecting to RabbitMQ, %s", rErr)
					}

					log.PrintfNoContext(rabbit.AppID, component, "RabbitMQ Reconnected after connection error")

					return
				}
			}
		}
	}()

}

func (rabbit *RabbitMq) CleanUp() error {

	for _, subChannel := range rabbit.subscriptionChannels {
		close(subChannel)
	}

	rabbit.subscriptionChannels = make(map[string]chan bool, 0)
	rabbit.registeredTopic = make(map[string]bool)

	if rabbit.publishChannel != nil {
		err := rabbit.publishChannel.Close()

		if err != nil {
			return fmt.Errorf("error closing public channel while cleaning up rabbitmq connection, %s", err)
		}
	}

	return nil
}

// RegisterTopic Should be called in the initialization to create an exchange
// If the exchange exists it's ignored
func (rabbit *RabbitMq) RegisterTopic(topic string) (err error) {

	channel, err := rabbit.MqConnection.Channel()

	if err != nil {
		return err
	}

	defer func() {
		err := channel.Close()

		if err != nil {
			log.ErrorfNoContext(rabbit.AppID, component, "Error closing channel while registering topic, %s", err)
		}
	}()

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

	log.PrintfNoContext(rabbit.AppID, component, "Registered topic %s for app %s", topic, rabbit.AppID)

	return nil
}

func (rabbit *RabbitMq) InitializeQueue(topic string) (err error) {

	for attempts := 1; attempts < 4; attempts++ {

		err = rabbit.declareQueue(topic)
		if err == nil {
			break
		}

		log.PrintfNoContext(rabbit.AppID, component, "Failed to initialize queue for topic %s. Waiting (30 sec) for next attempt. Total Attempts = %d", topic, attempts)

		time.Sleep(time.Second * 30)

	}

	if err != nil {
		log.PrintfNoContext(rabbit.AppID, component, "Failed to initialize queue for topic %s. %s", topic, err)
		return err
	}

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
func (rabbit *RabbitMq) declareQueue(topic string) (err error) {

	channel, err := rabbit.MqConnection.Channel()

	if err != nil {
		return err
	}

	defer func() {
		err := channel.Close()

		if err != nil {
			log.ErrorfNoContext(rabbit.AppID, component, "Error closing channel while initializing queue, %s", err)
		}
	}()

	appQueueName := formQueueName(rabbit.AppID, topic)
	deadLetterName := formDeadLetterName(rabbit.AppID, topic)

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

func (rabbit *RabbitMq) PublishToTopic(ctx context.Context, topic string, event []byte, contentType string) (err error) {

	appID := ctx.Value(appctx.AppIdHeader).(string)
	correlationID := ctx.Value(appctx.CorrelationIdHeader).(string)

	if !rabbit.registeredTopic[topic] {
		return fmt.Errorf("app %s is not registered for topic %s", appID, topic)
	}

	if rabbit.publishChannel == nil {

		err := rabbit.newPublishChannel()

		if err != nil {
			return err
		}

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
			AppId:         string(appID),
		})

	if err == amqp.ErrClosed {
		log.ErrorfNoContext(rabbit.AppID, component, "Error while publishing on channel or connection, %s. Retry open channel for publishing...", err)

		err := rabbit.newPublishChannel()

		if err != nil {
			return err
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
				AppId:         string(appID),
			})

		if err != nil {
			return err
		}

	}

	if err != nil && err != amqp.ErrClosed {
		return err
	}

	return nil
}

func (rabbit *RabbitMq) newPublishChannel() (err error) {

	channel, err := rabbit.MqConnection.Channel()

	if err != nil {
		return err
	}

	rabbit.publishChannel = channel

	log.PrintfNoContext(rabbit.AppID, component, "New channel set to publish channel.", err)

	return nil

}

func (rabbit *RabbitMq) SubscribeToTopic(topic string, processFunc ProcessEvent) (err error) {

	return rabbit.SubscribeToTopicWithMaxMsg(topic, processFunc, 0)
}

func (rabbit *RabbitMq) SubscribeToTopicWithMaxMsg(topic string, processFunc ProcessEvent, maxMessages int) (err error) {

	appQueueName := formQueueName(rabbit.AppID, topic)

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

				rabbit.handleDelivery(delivery, processFunc)

			case <-rabbit.subscriptionChannels[topic]:

				err := channel.Close()

				if err != nil {
					log.ErrorfNoContext(rabbit.AppID, component, "Error closing channel while ending subscription, %s", err)
				}

				return

			}
		}
	}()

	log.PrintfNoContext(rabbit.AppID, component, "App %s Subscribed to topic %s", rabbit.AppID, topic)

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
func (rabbit *RabbitMq) handleDelivery(delivery amqp.Delivery, processFunc ProcessEvent) {

	if delivery.CorrelationId == "" {
		id, _ := uuid.NewV4()
		delivery.CorrelationId = id.String()
	}

	ctx := appctx.NewContextFromDelivery(rabbit.AppID, delivery)

	err := processFunc(ctx, delivery.Body, delivery.ContentType)

	if err != nil {

		log.Errorf(ctx, component, "Error handling delivery, %s", err)

		if delivery.Redelivered {
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

func formQueueName(appID app.ApplicationID, topic string) string {
	return fmt.Sprintf("%s->%s", appID, topic)

}

func formDeadLetterName(appID app.ApplicationID, topic string) string {
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
