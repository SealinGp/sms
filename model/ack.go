package model

import "encoding/json"

type ACK struct {
	Key string `json:"key"`
}

func UnmarshalACK(data []byte) *ACK {
	msg := &ACK{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil
	}
	return msg
}
