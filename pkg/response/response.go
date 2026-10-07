package response

type Meta struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Meta Meta `json:"meta"`
	Data any  `json:"data"`
}

type ErrorResponse struct {
	Meta   Meta `json:"meta"`
	Data   any  `json:"data"`
	Errors any  `json:"errors"`
}

func Success(message string, data any) SuccessResponse {
	return SuccessResponse{
		Meta: Meta{
			Success: true,
			Message: message,
		},
		Data: data,
	}
}

func Error(message string, errors any) ErrorResponse {
	return ErrorResponse{
		Meta: Meta{
			Success: false,
			Message: message,
		},
		Data:   nil,
		Errors: errors,
	}
}

