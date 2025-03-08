package handlers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	tracer = otel.Tracer("demo-handlers")
	logger *zap.Logger
)

// InitLogger initializes the package logger with the shared logger instance
func InitLogger(l *zap.Logger) {
	logger = l.With(
		zap.String("component", "demo-handlers"),
	)
}

// SimulateWork adds random delay and occasionally returns errors
func SimulateWork(c echo.Context) error {
	_, span := tracer.Start(c.Request().Context(), "simulate-work")
	defer span.End()

	// Simulate random processing time
	duration := time.Duration(rand.Int63n(1000)) * time.Millisecond
	time.Sleep(duration)

	span.SetAttributes(attribute.Float64("processing.duration_ms", float64(duration.Milliseconds())))

	// Simulate random errors (10% chance)
	if rand.Float32() < 0.1 {
		err := fmt.Errorf("random processing error")
		span.RecordError(err)
		return c.JSON(500, map[string]interface{}{"error": err.Error()})
	}

	return c.JSON(200, map[string]interface{}{"duration_ms": duration.Milliseconds()})
}

// SimulateWorkError always returns an error
func SimulateWorkError(c echo.Context) error {
	_, span := tracer.Start(c.Request().Context(), "simulate-work-error")
	defer span.End()

	// Simulate random processing time before error
	duration := time.Duration(rand.Int63n(500)) * time.Millisecond
	time.Sleep(duration)

	err := fmt.Errorf("simulated work processing error")
	span.RecordError(err)
	span.SetAttributes(attribute.Float64("processing.duration_ms", float64(duration.Milliseconds())))

	return c.JSON(500, map[string]interface{}{"error": err.Error(), "duration_ms": duration.Milliseconds()})
}

// SimulateMemoryError simulates a memory allocation error
func SimulateMemoryError(c echo.Context) error {
	_, span := tracer.Start(c.Request().Context(), "simulate-memory-error")
	defer span.End()

	err := fmt.Errorf("out of memory error simulation")
	span.RecordError(err)
	span.SetAttributes(attribute.String("error.type", "memory_allocation_error"))

	return c.JSON(500, map[string]interface{}{"error": err.Error()})
}

// SimulateCPUError simulates a CPU overload error
func SimulateCPUError(c echo.Context) error {
	_, span := tracer.Start(c.Request().Context(), "simulate-cpu-error")
	defer span.End()

	err := fmt.Errorf("CPU overload error simulation")
	span.RecordError(err)
	span.SetAttributes(attribute.String("error.type", "cpu_overload"))

	return c.JSON(500, map[string]interface{}{"error": err.Error()})
}

// Global variable to simulate memory leak
var leakedMemory [][]byte

// SimulateMemoryLeak gradually increases memory usage
func SimulateMemoryLeak(c echo.Context) error {
	_, span := tracer.Start(c.Request().Context(), "simulate-memory-leak")
	defer span.End()

	// Create a slice that will stay in memory
	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i)
	}

	// Store in a global variable to prevent garbage collection
	leakedMemory = append(leakedMemory, data)

	span.SetAttributes(attribute.Int("memory.leaked_mb", len(leakedMemory)))

	return c.JSON(200, map[string]interface{}{"leaked_mb": len(leakedMemory)})
}

// SimulateCPULoad generates high CPU usage
func SimulateCPULoad(c echo.Context) error {
	_, span := tracer.Start(c.Request().Context(), "simulate-cpu-load")
	defer span.End()

	// Start CPU-intensive calculation
	result := 0.0
	for i := 0; i < 1000000; i++ {
		result += float64(i) * rand.Float64()
	}

	span.SetAttributes(attribute.Float64("cpu.result", result))

	return c.JSON(200, map[string]interface{}{"result": result})
}

// GenerateRandomLogs generates various types of logs
func GenerateRandomLogs(c echo.Context) error {
	_, span := tracer.Start(c.Request().Context(), "generate-random-logs")
	defer span.End()

	// Define log types and their messages
	logTypes := []struct {
		level   zapcore.Level
		message string
	}{
		{zapcore.InfoLevel, "User logged in successfully"},
		{zapcore.InfoLevel, "Payment processed"},
		{zapcore.InfoLevel, "Order created"},
		{zapcore.WarnLevel, "High API latency detected"},
		{zapcore.WarnLevel, "Database connection pool nearly full"},
		{zapcore.ErrorLevel, "Failed to process payment"},
		{zapcore.ErrorLevel, "Database connection timeout"},
		{zapcore.ErrorLevel, "Invalid API token"},
		{zapcore.DebugLevel, "Cache miss for key: user_123"},
		{zapcore.DebugLevel, "Processing batch job"},
	}

	// Generate random number of logs (1-5)
	numLogs := rand.Intn(5) + 1
	generatedLogs := make([]string, 0, numLogs)

	for i := 0; i < numLogs; i++ {
		// Pick random log type
		logType := logTypes[rand.Intn(len(logTypes))]

		// Add some random context
		userID := fmt.Sprintf("user_%d", rand.Intn(1000))
		requestID := fmt.Sprintf("req_%d", rand.Intn(10000))
		traceID := fmt.Sprintf("trace_%d", rand.Intn(10000))

		// Common fields for structured logging
		fields := []zapcore.Field{
			zap.String("user_id", userID),
			zap.String("request_id", requestID),
			zap.String("trace_id", traceID),
			zap.String("level", logType.level.String()),
			zap.Time("timestamp", time.Now()),
			zap.String("msg", logType.message),
		}

		// Log with proper level and structured fields
		switch logType.level {
		case zapcore.InfoLevel:
			logger.Info(logType.message, fields...)
		case zapcore.WarnLevel:
			logger.Warn(logType.message, fields...)
		case zapcore.ErrorLevel:
			logger.Error(logType.message, fields...)
		case zapcore.DebugLevel:
			logger.Debug(logType.message, fields...)
		}

		generatedLogs = append(generatedLogs, fmt.Sprintf("%s: %s", logType.level.String(), logType.message))

		// Add small delay between logs for better visibility
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	}

	span.SetAttributes(attribute.Int("logs.generated", numLogs))

	return c.JSON(200, map[string]interface{}{
		"logs_generated": numLogs,
		"logs":           generatedLogs,
	})
}
