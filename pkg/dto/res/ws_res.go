package res

type JsonResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Type    string      `json:"type,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

type Ws_UserResponse struct {
	UserID   uint `json:"user_id"`
	IsOnline bool `json:"is_online"`
}
