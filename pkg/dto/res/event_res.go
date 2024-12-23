package res

type EventPayload struct {
	Topic     string `json:"topic"`
	Action    string `json:"action"`
	Message   string `json:"message"`
	UserId    uint   `json:"user_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Timestamp string `json:"timestamp"`
}
