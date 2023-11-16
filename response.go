package httpecho

type HttpEchoResponse struct {
	StatusCode int                   `json:"status_code"`
	Hostname   string                `json:"hostname"`
	Path       string                `json:"path"`
	Method     string                `json:"method"`
	Query      map[string][]string   `json:"query"`
	Headers    map[string][]string   `json:"headers"`
	Body       *string               `json:"body,omitempty"`
	Json       *interface{}          `json:"json,omitempty"`
	Jwt        *HttpEchoResponse_Jwt `json:"jwt,omitempty"`
}

type HttpEchoResponse_Jwt struct {
	Header    map[string]interface{} `json:"header"`
	Payload   map[string]interface{} `json:"payload"`
	Signature string                 `json:"signature"`
}
