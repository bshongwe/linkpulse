module github.com/bshongwe/linkpulse/backend

go 1.23

require (
    github.com/google/uuid v1.6.0
    github.com/jackc/pgx/v5 v5.7.0
    github.com/jmoiron/sqlx v1.4.0
    github.com/spf13/viper v1.19.0
    go.uber.org/zap v1.27.0
    go.opentelemetry.io/otel v1.28.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.28.0
	go.opentelemetry.io/otel/sdk v1.28.0
	go.opentelemetry.io/otel/trace v1.28.0
)