package v8

import (
	"encoding/base64"
	"encoding/json"
)

func EncodeSsrHeaders(headers map[string]string) string {
	bytes, _ := json.Marshal(headers)
	if len(bytes) > 0 {
		return base64.RawURLEncoding.EncodeToString(bytes)
	}
	return ""
}

func DecodeSsrHeaders(s string) map[string]string {
	bytes, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil
	}
	var ret map[string]string
	err = json.Unmarshal(bytes, &ret)
	if err != nil {
		return nil
	}
	return ret
}
