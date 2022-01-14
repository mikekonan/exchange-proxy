package kucoin

import (
	"encoding/json"
)

//go:generate easyjson -lower_camel_case -omit_empty wire.go

//easyjson:json
type bulletPublicResponse struct {
	Code string `json:"code"`
	Data struct {
		Token           string `json:"token"`
		InstanceServers []struct {
			Endpoint     string `json:"endpoint"`
			Encrypt      bool   `json:"encrypt"`
			Protocol     string `json:"protocol"`
			PingInterval int64  `json:"pingInterval"`
			PingTimeout  int64  `json:"pingTimeout"`
		} `json:"instanceServers"`
	} `json:"data"`
}

//easyjson:json
type welcomeMessageResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

//easyjson:json
type pingMessageRequest struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

//easyjson:json
type subscribeMessageRequest struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	Topic          string `json:"topic"`
	PrivateChannel bool   `json:"privateChannel"`
	Response       bool   `json:"response"`
}

//easyjson:json
type kLineUpdateMessageEntry struct {
	Symbol  string `json:"symbol"`
	Candles kLine  `json:"candles"`
}

//easyjson:json
type genericMessageResponse struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Topic   string          `json:"topic"`
	Subject string          `json:"subject"`
	Data    json.RawMessage `json:"data"`
}

//easyjson:json
type kLinesResponse struct {
	Code    string `json:"code"`
	Klines  kLines `json:"data"`
	Message string `json:"message"`
}

//easyjson:json
type genericResponse struct {
	Code    string          `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

//easyjson:json
type kLine [7]string

//easyjson:json
type kLines []*kLine
