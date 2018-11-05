package io

import (
	"strings"
)

type Request struct {
	Command string `json:"cmd"`
	Action  string `json:"action"`
	Data    string `json:"data"`
}

func CreateRequest(msg string) *Request {
	if len(msg) == 0 {
		return nil
	}
	parts := strings.Split(msg, " ")
	req := &Request{}
	if len(parts) >= 1 {
		req.Command = strings.ToUpper(parts[0])
	}
	if len(parts) >= 2 {
		req.Action = strings.ToUpper(parts[1])
	}
	if len(parts) >= 3 {
		req.Data = strings.Join(parts[2:], " ")
	}

	return req
}
