package http

import (
	"context"
	"io"
	"log/slog"
	"net/http"
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
	audit(context.Context, <-chan auditEvent) error
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

	_ = context.AfterFunc(ctx, func() {
		if err := a.broadcast(ctx, a.c); err != nil {
			slog.Error( err.Error())
		}
	})

	req = req.WithContext(ctx)
	a.h.ServeHTTP(res, req)
	<-ctx.Done()
}

func (s *Server) newAuditedHandler(action string, h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		req = req.WithContext(ctx)
		s.auditChannels = append(s.auditChannels, make(chan auditEvent))
		_ = context.AfterFunc(ctx, func() {
			a := auditEvent{
				Action: action,
				TS:     time.Now().Unix(),
				UserID: nil,
				URL:    "",
			}
			l := len(s.auditChannels) - 1
			s.auditChannels[l] <- a
			slog.Debug(fmt.Sprintf("sent %+von s.auditChannels[l]", a))
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
