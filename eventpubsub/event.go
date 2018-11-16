package eventpubsub

import (
	"encoding/json"
	"fmt"
	"time"
)

type (

	AppEvent struct {
		EventType string
		Timestamp time.Time
		Data interface{}
	}

)

func NewAppEvent(eventType string, data interface{}) AppEvent{

	return AppEvent{
		EventType:eventType,
		Timestamp: time.Now().UTC(),
		Data: data,
	}
}

func (event *AppEvent) ToJSON() (eventJSON []byte, err error){

	eventJSON, err = json.Marshal(event)

	if err != nil {
		return eventJSON, fmt.Errorf("erro serializing event %s to JSON, %s", event.EventType, err)
	}

	return eventJSON, nil
}
