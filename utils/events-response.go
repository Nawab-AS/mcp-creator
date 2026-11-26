package utils

// events-response.go - simple wrapper to avoid JS Promise<Error>
// Example usage:
// 1. return utils.NewResponse(true, "<Error message>") // for error
// 2. return utils.NewResponse(false, "<Success message>") // for success

type Response struct {
	IsErr   bool
	Message string
}

func (u *Utils) NewResponse(isErr bool, message string) Response {
	return Response{
		IsErr:   isErr,
		Message: message,
	}
}
