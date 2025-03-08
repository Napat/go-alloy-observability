package middleware

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
)

var (
	tracer         = otel.Tracer("http-middleware")
	meter          = otel.Meter("http-middleware")
	requestCounter metric.Int64Counter
)

func init() {
	var err error
	requestCounter, err = meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		panic(err)
	}
}

// Telemetry middleware adds tracing and metrics to each request
func Telemetry() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			path := c.Request().URL.Path
			method := c.Request().Method

			// Extract tracing context from headers
			ctx := c.Request().Context()
			propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
			ctx = propagator.Extract(ctx, propagation.HeaderCarrier(c.Request().Header))

			// Start a new span
			ctx, span := tracer.Start(ctx, fmt.Sprintf("HTTP %s %s", method, path))
			defer span.End()

			// Add basic span attributes
			span.SetAttributes(
				attribute.String("http.method", method),
				attribute.String("http.path", path),
			)

			// Store context in request
			c.SetRequest(c.Request().WithContext(ctx))

			// Process request
			err := next(c)

			// Calculate request duration
			duration := time.Since(start).Milliseconds()

			// Record metrics with proper labels for Prometheus-style queries
			attrs := []attribute.KeyValue{
				attribute.String("method", method),
				attribute.String("path", path),
			}

			if err != nil {
				attrs = append(attrs, attribute.String("status", "error"))
			} else {
				attrs = append(attrs, attribute.String("status", "success"))
			}

			// Record request count
			requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

			// Add duration to span
			span.SetAttributes(attribute.Int64("http.duration_ms", duration))

			return err
		}
	}
}
