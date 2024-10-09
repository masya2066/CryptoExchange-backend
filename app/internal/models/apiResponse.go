package models

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func ResponseMsg(success bool, message string, code int) interface{} {
	if code == 0 {
		return SuccessResponse{
			Success: success,
			Message: message,
		}
	}

	return ErrorResponse{
		Success: success,
		Message: message,
		Code:    code,
	}
}
