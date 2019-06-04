package dto

// APIResponse is common response structure for all responses
type APIResponse struct {
	Data    interface{} `json:"data"`
	Success bool        `json:"success"`
}
