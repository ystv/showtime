package livestream

import (
	"encoding/json"
	"fmt"
	"time"
)

type (
	// EventType is the value of a livestream_events record's event_type field.
	// This type must stay in sync with the livestream_event_type enum in the database.
	EventType        string
	EventWithoutData struct {
		ID   int       `db:"livestream_event_id" json:"livestreamEventID"`
		Type EventType `db:"event_type" json:"type"`
		Time time.Time `db:"event_time" json:"time"`
	}
	// Event is a livestream_events record.
	Event struct {
		EventWithoutData
		Data EventPayload `db:"event_data" json:"data"`
	}
)

const (
	// EventStarted is when a livestream is started.
	EventStarted EventType = "started"
	// EventEnded is when a livestream is ended.
	EventEnded EventType = "ended"
	// EventLinked is when a livestream is linked to an integration.
	EventLinked EventType = "linked"
	// EventUnlinked is when a livestream is unlinked from an integration.
	EventUnlinked EventType = "unlinked"
	// EventStreamReceived is when ShowTime! starts receiving an incoming stream from nginx.
	EventStreamReceived EventType = "streamReceived"
	// EventStreamLost is when nginx reports it has stopped receiving a stream.
	EventStreamLost EventType = "streamLost"
	// EventError is when an error occurs while forwarding a stream.
	EventError EventType = "error"
)

// EventPayload is the type of all livestream event payloads, used only for type checking.
type EventPayload interface {
	isEventPayload()
}

// UnmarshalEventPayload unmarshals a JSON event payload into the appropriate type.
func UnmarshalEventPayload(typ EventType, raw json.RawMessage) (EventPayload, error) {
	var data EventPayload
	switch typ {
	case EventStarted:
		data = &EventStartedPayload{}
	case EventEnded:
		data = &EventEndedPayload{}
	case EventLinked:
		data = &EventLinkedPayload{}
	case EventUnlinked:
		data = &EventUnlinkedPayload{}
	case EventStreamReceived:
		data = &EventStreamReceivedPayload{}
	case EventStreamLost:
		data = &EventStreamLostPayload{}
	case EventError:
		data = &EventErrorPayload{}
	default:
		return nil, fmt.Errorf("unknown event type: %s", typ)
	}
	if err := json.Unmarshal(raw, data); err != nil {
		return nil, err
	}
	return data, nil
}

type EventStartedPayload struct{}

func (EventStartedPayload) isEventPayload() {}

type EventEndedPayload struct{}

func (EventEndedPayload) isEventPayload() {}

type EventLinkedPayload struct {
	IntegrationType IntegrationType `json:"integrationType"`
	IntegrationID   string          `json:"integrationID"`
}

func (EventLinkedPayload) isEventPayload() {}

type EventUnlinkedPayload struct {
	IntegrationType IntegrationType `json:"integrationType"`
	IntegrationID   string          `json:"integrationID"`
}

func (EventUnlinkedPayload) isEventPayload() {}

type EventStreamReceivedPayload struct{}

func (EventStreamReceivedPayload) isEventPayload() {}

type EventStreamLostPayload struct{}

func (EventStreamLostPayload) isEventPayload() {}

type EventErrorPayload struct {
	Err     string `json:"err"`
	Context string `json:"context"`
}

func (EventErrorPayload) isEventPayload() {}
