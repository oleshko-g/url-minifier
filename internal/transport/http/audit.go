package http

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
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
	subscribe(context.Context, <-chan auditEvent)
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
	idx := len(s.auditSubjects) - 1 // index of the appended audited handler subject
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		h(res, req)
		ctx := req.Context()
		userID, _ := userIDFromContext(ctx)
		originURL, _ := originalURLFromCtx(ctx)

		context.AfterFunc(ctx, func() {
			s.auditSubjects[idx] <- auditEvent{
				Action: action,
				TS:     time.Now().Unix(),
				UserID: &userID,
				URL:    originURL,
			}
		})
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

type auditFile struct {
	fp      *os.File
	mu      sync.Mutex
	enabled bool
	Source  string
}

func (a *auditFile) subscribe(ctx context.Context, auditEvents <-chan auditEvent) {
	for {
		select {
		case <-ctx.Done():
		case auditEvent := <-auditEvents:
			err := json.NewEncoder(a.fp).Encode(auditEvent)
			if err != nil {
				slog.Error(err.Error())
			}
		}
	}
}

// Set opens or creates the audit file or returns an error
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

type auditURL struct {
	url     *url.URL
	enabled bool
	Source  string
}

func (a *auditURL) subscribe(ctx context.Context, auditEvent <-chan auditEvent) {
	client := &http.Client{
		Timeout: time.Second,
	}

	for {
		select {
		case <-ctx.Done():
		case auditEvent := <-auditEvent:
			auditEventData, err := json.Marshal(auditEvent)
			if err != nil {
				slog.Error(err.Error())
			}
			r := bytes.NewReader(auditEventData)

			ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			req, err := http.NewRequestWithContext(ctx, "POST", a.url.String(), r)
			if err != nil {
				slog.Error(err.Error())
			}

			res, err := client.Do(req)
			if err != nil {
				slog.Error(err.Error())
			}
			res.Body.Close()
			cancel()
		}
	}
}

// Set parses s into a [url.URL] and sets it as the value of audit URL
func (a *auditURL) Set(s string) error {
	parsedURL, err := url.Parse(s)
	if err != nil {
		return err
	}

	a.url = parsedURL
	a.enabled = true
	return nil
}

// String return the opaque representation of an audit URL
func (a *auditURL) String() string {
	if a.url == nil {
		return ""
	}

	return a.url.Opaque
}
