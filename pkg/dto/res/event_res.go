package res

type EventPayload struct {
	Topic     string `json:"topic"`
	Action    string `json:"action"`
	Message   string `json:"message"`
	Payload   string `json:"payload"`
	CreatedAt string `json:"created_at"`
}
