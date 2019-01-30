package eventpubsub

import (
	"encoding/json"
	"fmt"
	"time"
)

type (
	AppEvent struct {
		EventType string
		Timestamp int64
		Data      json.RawMessage
	}
)

func NewAppEvent(eventType string, data json.RawMessage) AppEvent {

	return AppEvent{
		EventType: eventType,
		Timestamp: time.Now().UTC().Unix(),
		Data:      data,
	}
}

func NewAppEventWithTimestamp(eventType string, data json.RawMessage, timestamp int64) AppEvent {

	return AppEvent{
		EventType: eventType,
		Timestamp: timestamp,
		Data:      data,
	}
}

func NewAppEventFromJSON(event []byte) (appEvent AppEvent, err error) {

	err = json.Unmarshal(event, &appEvent)

	if err != nil {
		return appEvent, fmt.Errorf("erro deserializing event %s to JSON, %s", string(event), err)
	}

	return appEvent, nil

}

func (event *AppEvent) ToJSON() (eventJSON []byte, err error) {

	eventJSON, err = json.Marshal(event)

	if err != nil {
		return eventJSON, fmt.Errorf("erro serializing event %s to JSON, %s", event.EventType, err)
	}

	return eventJSON, nil
}
