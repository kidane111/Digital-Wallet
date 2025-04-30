package doc

type SuccessResponse struct {
	// OK is only true if the request was successful.
	OK bool `json:"ok"`
	// Data contains the actual data of the response.
	Data interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	// OK is only false if the request was not successful.
	OK bool `json:"ok"  default:"false"`
	// Error contains the error detail if the request was not successful.
	Error struct { // Code is the error code. It is not status code
		Code int `json:"code"`
		// Message is the error message.
		Message string `json:"message,omitempty"`
		// Description is the error description.
		Description string `json:"description,omitempty"`
		// StackTrace is the stack trace of the error.
		// It is only returned for debugging
		StackTrace string `json:"stack_trace,omitempty"`
		// FieldError is the error detail for each field, if available that is.
		FieldError []FieldError `json:"field_error,omitempty"`
	} `json:"error,omitempty"`
}
type FieldError struct {
	// Name is the name of the field that caused the error.
	Name string `json:"name"`
	// Description is the error description for this field.
	Description string `json:"description"`
}
