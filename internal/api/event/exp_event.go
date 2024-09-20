package event

type ExpEvent struct {
}

func NewNotificationEvent() *ExpEvent {
	return &ExpEvent{}
}

func (e *ExpEvent) Channel() string {
	return "exp"
}

func (e *ExpEvent) EventType() string {
	return "new-exp"
}

func (e *ExpEvent) Data() map[string]interface{} {
	return map[string]interface{}{
		"message": "",
	}
}
