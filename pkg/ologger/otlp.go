package ologger

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// OTLPWriter implements io.Writer interface for sending logs to OTLP endpoint
type OTLPWriter struct {
	provider *sdklog.LoggerProvider
}

// Write implements io.Writer interface
func (w *OTLPWriter) Write(p []byte) (n int, err error) {
	// Convert byte slice to string
	message := string(p)

	// Create logRecord
	logRecord := new(otellog.Record)
	logRecord.SetBody(otellog.StringValue(message))

	// Get the logger from the provider using the name "otlp"
	otelLogger := w.provider.Logger("otlp")

	// Emit the log message using Emit with the logRecord
	otelLogger.Emit(context.Background(), *logRecord)

	// Return the length of the byte slice as the number of bytes written
	return len(p), nil
}

// InitLogger initializes a new zap logger with OTLP integration
func InitLogger(ctx context.Context, serviceName, otlpEndpoint string, res *resource.Resource) (*zap.Logger, *sdklog.LoggerProvider, error) {
	// Initialize logs exporter
	logExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint(otlpEndpoint),
		otlploghttp.WithURLPath("/v1/logs"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create log exporter: %v", err)
	}

	processor := sdklog.NewBatchProcessor(logExporter)

	logProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(processor),
		sdklog.WithResource(res),
	)

	// Create production encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "timestamp",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			switch l {
			case zapcore.DebugLevel:
				enc.AppendString("debug")
			case zapcore.InfoLevel:
				enc.AppendString("info")
			case zapcore.WarnLevel:
				enc.AppendString("warn")
			case zapcore.ErrorLevel:
				enc.AppendString("error")
			default:
				enc.AppendString(l.String())
			}
		},
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create OTLP hook for Zap
	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(&OTLPWriter{logProvider}),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		),
	)

	// Create base logger with OTLP hook
	baseLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// Set service name in logger
	logger := baseLogger.With(zap.String("service.name", serviceName))

	return logger, logProvider, nil
}
