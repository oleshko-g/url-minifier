package http

type auditEvent struct {
	TS     int64   `json:"ts"`
	Action string  `json:"action"`
	UserID *string `json:"user_id,omitempty"`
	URL    string  `json:"url"`
}
