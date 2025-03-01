package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // เพิ่ม pprof handler เพื่อให้ Alloy ดึงข้อมูลได้
	"os"
	"runtime"
	"time"

	"demo-observability/internal/config"
	"demo-observability/internal/handlers"
	"demo-observability/internal/middleware"
	"demo-observability/pkg/ologger"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.uber.org/zap"
)

var (
	logger      *zap.Logger
	tracer      *trace.TracerProvider
	meter       *metric.MeterProvider
	logProvider *sdklog.LoggerProvider
	cfg         *config.Config
)

func initTelemetry() {
	// Initialize config
	var err error
	cfg, err = config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Create resource with service information
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("test-service"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// Initialize logs exporter
	ctx := context.Background()
	otlpEndpoint := fmt.Sprintf("%s:%s", cfg.Services.OTLP.Host, cfg.Services.OTLP.Port)

	logger, logProvider, err = ologger.InitLogger(ctx, "test-service", otlpEndpoint, res)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	// Share logger with handlers
	handlers.InitLogger(logger)

	// Initialize tracer
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(fmt.Sprintf("%s:%s", cfg.Services.OTLP.Host, cfg.Services.OTLP.Port)))
	if err != nil {
		log.Printf("failed to create trace exporter: %v", err)
		os.Exit(1)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)
	tracer = tracerProvider

	// Initialize metrics
	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(fmt.Sprintf("%s:%s", cfg.Services.OTLP.Host, cfg.Services.OTLP.Port)))
	if err != nil {
		log.Printf("failed to create metric exporter: %v", err)
		os.Exit(1)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(2*time.Second))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	meter = meterProvider

	// Enable profiling
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)
}

func main() {
	ctx := context.Background()
	initTelemetry()
	defer logger.Sync()
	defer func() {
		if err := tracer.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown tracer provider: %v", err)
		}
	}()
	defer func() {
		if err := meter.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown meter provider: %v", err)
		}
	}()
	defer func() {
		if err := logProvider.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown log provider: %v", err)
		}
	}()

	e := echo.New()
	// Add telemetry middleware
	e.Use(middleware.Telemetry())

	// Demo endpoints that generate telemetry data
	e.GET("/ping", func(c echo.Context) error {
		logger.Info("received ping request")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "pong",
		})
	})
	e.GET("/demo/work", handlers.SimulateWork)
	e.GET("/demo/memory", handlers.SimulateMemoryLeak)
	e.GET("/demo/cpu", handlers.SimulateCPULoad)
	e.GET("/demo/logs", handlers.GenerateRandomLogs)

	// Start main server
	if err := e.Start(fmt.Sprintf(":%s", cfg.Services.Server.Port)); err != nil {
		logger.Fatal("server failed", zap.Error(err))
		os.Exit(1)
	}
}
