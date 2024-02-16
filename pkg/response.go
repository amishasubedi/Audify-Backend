package response

type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewErrorResponse(code int, message string, data interface{}) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
