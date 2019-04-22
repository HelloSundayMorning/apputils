package eventpubsub

import (
	"fmt"
	"github.com/HelloSundayMorning/apputils/appctx"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
)

func TestRabbitMq_RegisterTopic(t *testing.T) {

	const topic = "testRegisterTopic"
	rb, _ := NewRabbitMq("rabbitmq", "rabbitmq", "localhost")

	err := rb.RegisterTopic("testApp", topic)

	assert.Nil(t, err)
	assert.True(t, rb.registeredTopic[topic])

	ch, _ := rb.MqConnection.Channel()
	defer ch.Close()

	err = ch.ExchangeDeclarePassive(
		topic,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	assert.Nil(t, err)

	ch.ExchangeDelete(topic, false, true)

	rb.CleanUp()

}

func TestRabbitMq_InitializeQueue(t *testing.T) {

	const topic, appID = "testInitializeQueue", "testApp"
	expQueueName := fmt.Sprintf("%s->%s", appID, topic)
	expDeadQueueName := fmt.Sprintf("%s->%s.deadletter", appID, topic)

	rb, _ := NewRabbitMq("rabbitmq", "rabbitmq", "localhost")

	rb.RegisterTopic(appID, topic)

	err := rb.InitializeQueue(appID, topic)

	assert.Nil(t, err)

	ch, _ := rb.MqConnection.Channel()
	defer ch.Close()

	_, err = ch.QueueInspect(expQueueName)

	assert.Nil(t, err)

	_, err = ch.QueueInspect(expDeadQueueName)

	assert.Nil(t, err)

	ch.QueueDelete(expDeadQueueName, false, false, true)
	ch.QueueDelete(expQueueName, false, false, true)
	ch.ExchangeDelete(topic, false, true)
	ch.ExchangeDelete(expDeadQueueName, false, true)

	rb.CleanUp()

}

func TestRabbitMq_Publish(t *testing.T) {

	const topic, appID, event = "testPublish", "testApp", "testEvent"
	rb, _ := NewRabbitMq("rabbitmq", "rabbitmq", "localhost")

	ctx := appctx.NewContextFromValues(appID, "")

	err := rb.Publish(ctx, topic, []byte(event), "text/plain")

	assert.NotNil(t, err)
	assert.Equal(t, "app testApp is not registered for topic testPublish", err.Error())

	rb.RegisterTopic("testApp", topic)

	err = rb.Publish(ctx, topic, []byte(event), "text/plain")

	assert.Nil(t, err)

	ch, _ := rb.MqConnection.Channel()
	defer ch.Close()

	ch.ExchangeDelete(topic, false, true)

	rb.CleanUp()
}

func TestRabbitMq_Subscribe(t *testing.T) {

	const topic, appID, event = "testSubscribe", "testApp", "testEvent"
	expQueueName := fmt.Sprintf("%s->%s", appID, topic)
	expDeadQueueName := fmt.Sprintf("%s->%s.deadletter", appID, topic)
	rb, _ := NewRabbitMq("rabbitmq", "rabbitmq", "localhost")

	ch, _ := rb.MqConnection.Channel()
	defer ch.Close()

	ctx := appctx.NewContextFromValues(appID, "")


	err := rb.Subscribe(appID, topic, func(ctx context.Context, event []byte, contentType string) error {
		return nil
	})

	assert.NotNil(t, err)
	assert.Equal(t, "Exception (404) Reason: \"NOT_FOUND - no queue 'testApp->testSubscribe' in vhost '/'\"", err.Error())

	rb.RegisterTopic(appID, topic)
	rb.InitializeQueue(appID, topic)

	rChan := make(chan []byte)

	err = rb.Subscribe(appID, topic, func(ctx context.Context, event []byte, contentType string) error {

		rChan <- event

		return nil
	})

	assert.Nil(t, err)

	rb.Publish(ctx, topic, []byte(event), "")

	result := <- rChan

	assert.Equal(t, event, string(result))

	ch.QueueDelete(expDeadQueueName, false, false, true)
	ch.QueueDelete(expQueueName, false, false, true)
	ch.ExchangeDelete(topic, false, true)
	ch.ExchangeDelete(expDeadQueueName, false, true)

	rb.CleanUp()


}
