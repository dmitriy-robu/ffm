package event

import (
	"github.com/pusher/pusher-http-go/v5"
	"go-fitness/external/logger/sl"
	"log/slog"
)

type WebSocketClient struct {
	log  *slog.Logger
	conn pusher.Client
}

type WSInterface interface {
	TriggerEvent(e Event) error
}

type Event interface {
	Channel() string
	EventType() string
	Data() map[string]interface{}
}

func NewPusherEvent(
	log *slog.Logger,
	conn pusher.Client,
) *WebSocketClient {
	return &WebSocketClient{
		log:  log,
		conn: conn,
	}
}

func (w *WebSocketClient) TriggerEvent(e Event) error {
	const op = "event.PusherEvent.TriggerEvent"

	log := w.log.With(
		sl.String("op", op),
		sl.String("channel", e.Channel()),
		sl.String("event_type", e.EventType()),
		sl.Any("data", e.Data()),
	)

	if err := w.conn.Trigger(e.Channel(), e.EventType(), e.Data()); err != nil {
		log.Error("failed to trigger event", sl.Err(err))
		return nil
	}

	log.Info("event triggered")

	return nil
}
