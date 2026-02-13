package http

import (
	"context"
	"io"
	"net/http"
)

type auditEvent struct {
	TS     int64   `json:"ts"`
	Action string  `json:"action"`
	UserID *string `json:"user_id,omitempty"`
	URL    string  `json:"url"`
}

type subscriber interface {
	subscribe(context.Context, <-chan auditEvent) error
	io.Writer
}

type publisher interface {
	publish(action string) (chan auditEvent, error)
}

type broadcaster interface {
	broadcast(context.Context, chan<- auditEvent) error
	http.ResponseWriter
}
