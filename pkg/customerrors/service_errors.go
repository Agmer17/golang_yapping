package customerrors

type ServiceErrors struct {
	Code    int
	Message string
}

func (e *ServiceErrors) Error() string {
	return e.Message
}

func New(code int, message string) *ServiceErrors {
	return &ServiceErrors{
		Code:    code,
		Message: message,
	}
}
