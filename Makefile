# Makefile for Observability Stack

.PHONY: help
# Default target
help:
	@echo "Available commands:"
	@echo "  make up              		- Start all services"
	@echo "  make down            		- Stop all services"
	@echo "  make restart         		- Restart all services"
	@echo "  make logs            		- View logs from all services"
	@echo "  make logs-alloy      		- View logs from Alloy service"
	@echo "  make logs-loki       		- View logs from Loki service"
	@echo "  make logs-grafana    		- View logs from Grafana service"
	@echo "  make alloy-reload    		- Reload Alloy service"
	@echo "  make clean           		- Remove all data and containers"
	@echo "  make test-alloy-logs		- Send test logs to Alloy"
	@echo "  make test-alloy-metrics	- Send test counter/guage metrics to Alloy"
	@echo "  make test-alloy-metrics-sum-counter	- Send test counter metrics to Alloy"
	@echo "  make test-alloy-metrics-gauge	- Send test gauge metrics to Alloy"
	@echo "  make test-go-load       		- Generate test load on the application"
	@echo "  make test-go-random-logs 	- Generate random logs via golang endpoint"
	@echo "  make help            		- Show this help message"

.PHONY: up
# Start all services
up:
	docker compose up -d --build

.PHONY: down
# Stop all services
down:
	docker compose down

.PHONY: restart
# Restart all services
restart:
	docker compose restart

.PHONY: logs
# View logs from all services
logs:
	docker compose logs -f

.PHONY: logs-alloy
# View logs from specific services
logs-alloy:
	docker compose logs -f alloy

.PHONY: logs-loki
# View logs from Loki service
logs-loki:
	docker compose logs -f loki

.PHONY: logs-grafana
# View logs from Grafana service
logs-grafana:
	docker compose logs -f grafana

.PHONY: alloy-reload
# Reload Alloy service
alloy-reload:
	curl -X POST http://localhost:12345/-/reload

.PHONY: clean
# Clean up everything
clean:
	docker compose down -v
	rm -rf ./docker/data/*

.PHONY: test-alloy-logs
# Send test logs to Alloy
test-alloy-logs:
	curl -X POST http://localhost:4318/v1/logs \
	-H "Content-Type: application/json" \
	-d '{ \
	  "resourceLogs": [{ \
	    "resource": { \
	      "attributes": [{ \
	        "key": "service.name", \
	        "value": { "stringValue": "test-service" } \
	      }] \
	    }, \
	    "scopeLogs": [{ \
	      "logRecords": [{ \
	        "timeUnixNano": "'$$(date +%s000000000)'", \
	        "severityText": "INFO", \
	        "body": { \
	          "stringValue": "Test log message from Makefile" \
	        } \
	      }] \
	    }] \
	  }] \
	}'
	@echo "\nTest log sent. Check Grafana at http://localhost:13000"
	@echo "Use query: {service_name=\"test-service\"}"

.PHONY: test-alloy-metrics
# Send test metrics to Alloy
test-alloy-metrics:
	curl -X POST http://localhost:4318/v1/metrics \
	-H "Content-Type: application/json" \
	-d '{ \
	  "resourceMetrics": [{ \
	    "resource": { \
	      "attributes": [{ \
	        "key": "service.name", \
	        "value": { "stringValue": "test-service" } \
	      }] \
	    }, \
	    "scopeMetrics": [{ \
	      "metrics": [{ \
	        "name": "test.counter", \
	        "sum": { \
	          "dataPoints": [{ \
	            "asInt": "1", \
	            "timeUnixNano": "'$$(date +%s000000000)'" \
	          }], \
	          "isMonotonic": true, \
	          "aggregationTemporality": 1 \
	        } \
	      }, \
	      { \
	        "name": "test.gauge", \
	        "gauge": { \
	          "dataPoints": [{ \
	            "asDouble": "42.5", \
	            "timeUnixNano": "'$$(date +%s000000000)'" \
	          }] \
	        } \
	      }] \
	    }] \
	  }] \
	}'
	@echo "\nTest metrics sent. Check Grafana at http://localhost:13000"
	@echo "Use query: {__name__=~\".+\"}"
	@echo "Use query: test_gauge"
	@echo "Use query: rate(test_counter[5m])"

.PHONY: test-alloy-metrics-gauge
# Send test gauge metrics to Alloy
# consumeMetric() https://github1s.com/grafana/alloy/blob/main/internal/component/otelcol/exporter/prometheus/internal/convert/convert.go#L280-L281
# consumeGauge() https://github1s.com/grafana/alloy/blob/main/internal/component/otelcol/exporter/prometheus/internal/convert/convert.go#L302-L303
test-alloy-metrics-gauge:
	curl -X POST http://localhost:4318/v1/metrics \
	-H "Content-Type: application/json" \
	-d '{ \
		"resourceMetrics": [{ \
		"resource": {"attributes": [{"key": "service.name", "value": {"stringValue": "test-service"}}]}, \
		"scopeMetrics": [{ \
			"metrics": [{ \
			"name": "test.gauge", \
			"gauge": { \
				"dataPoints": [{"asDouble": "42.5", "timeUnixNano": "'$$(date +%s000000000)'"}] \
			} \
			}] \
		}] \
	}] \
	}'
	@echo "\nTest gauge metric sent. Check Grafana at http://localhost:13000"
	@echo "Use query: {__name__=~\".+\"}"
	@echo "Use query: test_gauge"

.PHONY: test-alloy-metrics-sum-counter
# Send test counter metrics to Alloy : please read 
# consumeMetric() https://github1s.com/grafana/alloy/blob/main/internal/component/otelcol/exporter/prometheus/internal/convert/convert.go#L280-L281
# consumeSum() https://github1s.com/grafana/alloy/blob/main/internal/component/otelcol/exporter/prometheus/internal/convert/convert.go#L413-L414
test-alloy-metrics-sum-counter:
	curl -X POST http://localhost:4318/v1/metrics \
	-H "Content-Type: application/json" \
	-d '{ \
	  "resourceMetrics": [{ \
	    "resource": { \
	      "attributes": [{ \
	        "key": "service.name", \
	        "value": { "stringValue": "test-service" } \
	      },{ \
	        "key": "job", \
	        "value": { "stringValue": "test-service" } \
	      }] \
	    }, \
	    "scopeMetrics": [{ \
	      "metrics": [{ \
	        "name": "test.counter", \
	        "sum": { \
	          "dataPoints": [ \
	            {"as_double": 600, "start_time_unix_nano": 1000000000, "timeUnixNano": "'$$(($$(date +%s)000000000 - 30000000000))'"}, \
				{"as_double": 700, "start_time_unix_nano": 1000000000, "timeUnixNano": "'$$(($$(date +%s)000000000 - 20000000000))'"}, \
				{"as_double": 800, "start_time_unix_nano": 1000000000, "timeUnixNano": "'$$(($$(date +%s)000000000 - 10000000000))'"}, \
	            {"as_double": 900, "start_time_unix_nano": 1000000010, "timeUnixNano": "'$$(date +%s000000000)'"} \
	          ], \
	          "isMonotonic": true, \
	          "aggregationTemporality": "AGGREGATION_TEMPORALITY_CUMULATIVE" \
	        } \
	      }] \
	    }] \
	  }] \
	}'
	@echo "\nTest counter metrics sent. Check Grafana at http://localhost:13000"
	@echo "Use query: {__name__=~\".+\"}"
	@echo "Use query: test_counter_total{job=\"test-service\"}"
	@echo "Use query: rate(test_counter_total{job="test-service"}[1m])"

.PHONY: test-alloy-metrics-sum-guage
# Send test guage metrics to Alloy : please read 
# consumeMetric() https://github1s.com/grafana/alloy/blob/main/internal/component/otelcol/exporter/prometheus/internal/convert/convert.go#L280-L281
# consumeSum() https://github1s.com/grafana/alloy/blob/main/internal/component/otelcol/exporter/prometheus/internal/convert/convert.go#L413-L414
test-alloy-metrics-sum-guage:
	curl -X POST http://localhost:4318/v1/metrics \
	-H "Content-Type: application/json" \
	-d '{ \
	  "resourceMetrics": [{ \
	    "resource": { \
	      "attributes": [{ \
	        "key": "service.name", \
	        "value": { "stringValue": "test-service" } \
	      },{ \
	        "key": "job", \
	        "value": { "stringValue": "test-service" } \
	      }] \
	    }, \
	    "scopeMetrics": [{ \
	      "metrics": [{ \
	        "name": "test.counter", \
	        "sum": { \
	          "dataPoints": [ \
	            {"as_double": 19, "start_time_unix_nano": 1000000000, "timeUnixNano": "'$$(date +%s000000000)'"} \
	          ], \
	          "isMonotonic": false, \
	          "aggregationTemporality": "AGGREGATION_TEMPORALITY_CUMULATIVE" \
	        } \
	      }] \
	    }] \
	  }] \
	}'
	@echo "\nTest counter metrics sent. Check Grafana at http://localhost:13000"
	@echo "Use query: {__name__=~\".+\"}"
	@echo "Use query: rate(test_counter[1m])"

.PHONY: test-alloy-metrics-sum-prometheus
# Send test prometheus metrics to Alloy : please read 
# consumeMetric() https://github1s.com/grafana/alloy/blob/main/internal/component/otelcol/exporter/prometheus/internal/convert/convert.go#L280-L281
# consumeSum() https://github1s.com/grafana/alloy/blob/main/internal/component/otelcol/exporter/prometheus/internal/convert/convert.go#L413-L414
# Alloy code: // Drop non-cumulative summaries for now,
test-alloy-metrics-sum-prometheus:
		curl -X POST http://localhost:4318/v1/metrics \
	-H "Content-Type: application/json" \
	-d '{ \
	  "resourceMetrics": [{ \
	    "resource": { \
	      "attributes": [{ \
	        "key": "service.name", \
	        "value": { "stringValue": "test-service" } \
	      },{ \
	        "key": "job", \
	        "value": { "stringValue": "test-service" } \
	      }] \
	    }, \
	    "scopeMetrics": [{ \
	      "metrics": [{ \
	        "name": "test.counter", \
	        "sum": { \
	          "dataPoints": [ \
	            {"as_double": 40, "start_time_unix_nano": 1000000000, "timeUnixNano": "'$$(date +%s000000000)'"} \
	          ], \
	          "isMonotonic": true, \
	          "aggregationTemporality": "AGGREGATION_TEMPORALITY_DELTA" \
	        } \
	      }] \
	    }] \
	  }] \
	}'
	@echo "\nTest counter metrics sent. Check Grafana at http://localhost:13000"
	@echo "Use query: {__name__=~\".+\"}"
	@echo "Use query: {__name__=~\"test_counter\"}"
	@echo "Use query: rate(test_counter[5m])"

.PHONY: test-alloy-traces
# Send test traces to Alloy
test-alloy-traces:
	curl -X POST http://localhost:4318/v1/traces \
	-H "Content-Type: application/json" \
	-d '{ \
	  "resourceSpans": [{ \
	    "resource": { \
	      "attributes": [{ \
	        "key": "service.name", \
	        "value": { "stringValue": "test-service" } \
	      }] \
	    }, \
	    "scopeSpans": [{ \
	      "spans": [{ \
	        "traceId": "'$$(openssl rand -hex 16)'", \
	        "spanId": "'$$(openssl rand -hex 8)'", \
	        "name": "test-operation", \
	        "kind": 1, \
	        "startTimeUnixNano": "'$$(date +%s000000000)'", \
	        "endTimeUnixNano": "'$$(date +%s999999999)'", \
	        "attributes": [{ \
	          "key": "http.method", \
	          "value": { "stringValue": "GET" } \
	        }] \
	      }] \
	    }] \
	  }] \
	}'
	@echo "\nTest trace sent. Check Grafana at http://localhost:13000"
	@echo "Go to Explore -> Tempo"
	
.PHONY: test-go-random-logs
# Generate random logs via golang endpoint
test-go-random-logs:
	@echo "Generating random logs..."
	@for i in {1..30}; do \
		curl -s "http://localhost:18080/demo/logs" > /dev/null & \
		sleep 0.2; \
	done
	@echo "Random log generation completed. Check Grafana -> Explore -> Loki"
	@echo "Use query: {service_name=\"test-service\"} | json"
	@echo "Use query: {service_name=\"test-service\"} | json | detected_level=~\"error|warn|info|debug\""

.PHONY: test-go-load
# Generate test load (random sleep 0.5 to 2 sec)
test-go-load:
	@echo "Generating load on application endpoints..."
	@echo "Start time: $$(date '+%Y-%m-%d %H:%M:%S')"
	@for i in {1..120}; do \
		curl -s "http://localhost:18080/demo/work" > /dev/null & \
		curl -s "http://localhost:18080/demo/work-error" > /dev/null & \
		curl -s "http://localhost:18080/demo/cpu" > /dev/null & \
		curl -s "http://localhost:18080/demo/cpu-error" > /dev/null & \
		curl -s "http://localhost:18080/demo/memory" > /dev/null & \
		curl -s "http://localhost:18080/demo/memory-error" > /dev/null; \
		sleep $$(echo "0.5 + $$RANDOM/32767*1.5" | bc -l); \
	done
	@echo "Load test completed at: $$(date '+%Y-%m-%d %H:%M:%S')"

.PHONY: generate-go-load
generate-go-load: test-go-random-logs test-go-load
