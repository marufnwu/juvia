package socket

// Request represents a JSON-RPC request from Panel to Agent.
type Request struct {
	ID     string      `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

// Response represents a JSON-RPC response from Agent to Panel.
type Response struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *ErrorObj   `json:"error,omitempty"`
}

// ErrorObj represents a JSON-RPC error.
type ErrorObj struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewResponse creates a successful response.
func NewResponse(id string, result interface{}) *Response {
	return &Response{ID: id, Result: result}
}

// NewErrorResponse creates an error response.
func NewErrorResponse(id, code, message string) *Response {
	return &Response{
		ID:    id,
		Error: &ErrorObj{Code: code, Message: message},
	}
}
