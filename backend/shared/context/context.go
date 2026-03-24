package ctxkey

type key string

const (
	TraceID     key = "trace_id"
	UserID      key = "user_id"
	WorkspaceID key = "workspace_id"
)
