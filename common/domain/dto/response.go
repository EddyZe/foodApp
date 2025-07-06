package dto

type Response struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  interface{} `json:"error,omitempty"`
}

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

type Message struct {
	Message string `json:"message"`
}
