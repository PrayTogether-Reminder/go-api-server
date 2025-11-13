package logger

import (
	"context"
	"log/slog"
)

const (
	// MemberIDKey is the context key for memberID (same as shared/context.MemberIDKey)
	memberIDKey = "member_id"
)

// ContextHandler automatically extracts memberID from context
type ContextHandler struct {
	slog.Handler
}

func NewContextHandler(h slog.Handler) *ContextHandler {
	return &ContextHandler{Handler: h}
}

// slog call this method
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Extract memberID from context if exists
	if memberID := ctx.Value(memberIDKey); memberID != nil {
		r.AddAttrs(slog.Any("memberID", memberID))
	}

	return h.Handler.Handle(ctx, r)
}
