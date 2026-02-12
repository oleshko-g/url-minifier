package http

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type auditEvent struct {
	TS     int64   `json:"ts"`
	Action string  `json:"action"`
	UserID *string `json:"user_id,omitempty"`
	URL    string  `json:"url"`
}

type observer interface {
	subscribe(context.Context, *sync.Cond) error
	write(context.Context, *auditEvent) error
}

type subject interface {
	publish(action string, userID *uuid.UUID) *sync.Cond
}
