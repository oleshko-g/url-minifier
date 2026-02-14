package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

type auditEvent struct {
	TS     int64   `json:"ts"`
	Action string  `json:"action"`
	UserID *string `json:"user_id,omitempty"`
	URL    string  `json:"url"`
}

const contextKeyOriginalURL contextKey = 2

func originalURLFromCtx(ctx context.Context) (url string, ok bool) {
	url, ok = ctx.Value(contextKeyOriginalURL).(string)
	return url, ok
}

type auditor interface {
	subscribe(context.Context, <-chan auditEvent) error
	io.Writer
}

type auditHandler struct {
	h        http.Handler
	action   string
	auditors []auditor
	c        chan<- auditEvent
}

func (a auditHandler) ServerHTTP(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	req = req.WithContext(ctx)
	_ = context.AfterFunc(ctx, func() {
		if err := a.broadcast(ctx, a.c); err != nil {
			slog.Error(err.Error())
		}
	})

	a.h.ServeHTTP(res, req)
}

func (s *Server) newAuditedHandler(action string, h http.HandlerFunc) http.HandlerFunc {
	s.auditSubjects = append(s.auditSubjects, make(chan auditEvent))
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		req = req.WithContext(ctx)
		userID, _ := userIDFromContext(ctx)
		originURL, _ := originalURLFromCtx(ctx)
		_ = context.AfterFunc(ctx, func() {
			a := auditEvent{
				Action: action,
				TS:     time.Now().Unix(),
				UserID: &userID,
				URL:    originURL,
			}
			l := len(s.auditSubjects) - 1
			s.auditSubjects[l] <- a
			slog.Info(fmt.Sprintf("sent %+von s.auditChannels[l]", a))
		})
		h(res, req)
	})
}

func (a auditHandler) broadcast(ctx context.Context, c chan<- auditEvent) error {
	userID, _ := userIDFromContext(ctx)
	select {
	case <-ctx.Done():
		return nil
	case c <- auditEvent{Action: a.action, TS: time.Now().Unix(), UserID: &userID, URL: ""}:
	}
	return nil
}

type auditSubject interface {
	register(action string, h http.Handler) (chan auditEvent, error)
	broadcast(context.Context, chan<- auditEvent) error
}

type auditFile struct {
	fp      *os.File
	mu      sync.Mutex
	enabled bool
	Source  string
}

func (a *auditFile) subscribe(ctx context.Context, channel <-chan auditEvent) error {
	for {
		select {
		case <-ctx.Done():
		case v := <-channel:
			err := json.NewEncoder(a.fp).Encode(v)
			if err != nil {
				slog.Error(err.Error())
			}
		}
	}
}

func (a *auditFile) Write(b []byte) (int, error) {
	return a.fp.Write(b)
}

// Set oprn or creates the audit file or returns an error
func (a *auditFile) Set(s string) error {
	// Write Read _, Read _ _, Read _ _
	const filePerm os.FileMode = 0o644

	fp, err := os.OpenFile(s, os.O_RDWR|os.O_CREATE|os.O_APPEND, filePerm)
	if err != nil {
		return err
	}

	a.fp = fp
	a.enabled = true
	return nil
}

// String returns the name of the audit file
func (a *auditFile) String() string {
	if a.fp == nil {
		return ""
	}

	return a.fp.Name()
}
