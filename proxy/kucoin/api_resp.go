package kucoin

import "encoding/json"

type apiResp struct {
	Code    string          `json:"code"`
	RawData json.RawMessage `json:"data,omitempty"`
	Message string          `json:"msg,omitempty"`
}

func (resp *apiResp) json() []byte {
	data, _ := json.Marshal(resp)
	return data
}


