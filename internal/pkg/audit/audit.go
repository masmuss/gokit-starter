// Package audit provides structured audit logging for security events.
package audit

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Logger writes structured audit events.
type Logger struct {
	log *slog.Logger
}

// New creates an audit logger.
func New(log *slog.Logger) *Logger {
	return &Logger{log: log}
}

// Info logs an audit entry at info level.
func (l *Logger) Info(event, outcome string, attrs ...slog.Attr) {
	l.logAttrs(slog.LevelInfo, event, outcome, attrs)
}

// Warn logs an audit entry at warn level.
func (l *Logger) Warn(event, outcome string, attrs ...slog.Attr) {
	l.logAttrs(slog.LevelWarn, event, outcome, attrs)
}

func (l *Logger) logAttrs(level slog.Level, event, outcome string, attrs []slog.Attr) {
	all := make([]slog.Attr, 0, len(attrs)+2)
	all = append(all, slog.String("event", event), slog.String("outcome", outcome))
	all = append(all, attrs...)
	l.log.LogAttrs(context.Background(), level, "audit", all...)
}

// UserID returns a user_id attribute.
func UserID(id uuid.UUID) slog.Attr {
	return slog.String("user_id", id.String())
}

// OrgID returns an org_id attribute.
func OrgID(id uuid.UUID) slog.Attr {
	return slog.String("org_id", id.String())
}

// IP extracts and returns a remote IP attribute from a request.
func IP(r *http.Request) slog.Attr {
	return slog.String("ip", clientIP(r))
}

// Email returns an email attribute.
func Email(email string) slog.Attr {
	return slog.String("email", email)
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if before, _, found := strings.Cut(xff, ","); found {
			return strings.TrimSpace(before)
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
