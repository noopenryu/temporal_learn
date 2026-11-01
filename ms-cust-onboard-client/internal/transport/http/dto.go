package http

type StartRequest struct {
	CustomerID  string `json:"customer_id,omitempty"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type ResumeRequest struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type CancelRequest struct {
	Reason string `json:"reason"`
}

type ErrorResponse struct {
	Error       string            `json:"error"`
	FieldErrors map[string]string `json:"field_errors,omitempty"`
}
