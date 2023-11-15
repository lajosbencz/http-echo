package httpecho

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

func ParseJwtString(jwt string) (*HttpEchoResponse_Jwt, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed JWT: %q", jwt)
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	headerStruct := map[string]interface{}{}
	if err := json.Unmarshal(headerBytes, &headerStruct); err != nil {
		return nil, err
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	payloadStruct := map[string]interface{}{}
	if err := json.Unmarshal(payloadBytes, &payloadStruct); err != nil {
		return nil, err
	}
	return &HttpEchoResponse_Jwt{
		Header:    headerStruct,
		Payload:   payloadStruct,
		Signature: parts[2],
	}, nil
}
