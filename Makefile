# Makefile for Observability Stack

.PHONY: up down restart logs logs-alloy logs-loki logs-grafana alloy-reload clean test-alloy-logs help test-load test-go-random-logs

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
	@echo "  make test-load       		- Generate test load on the application"
	@echo "  make test-go-random-logs 	- Generate random logs via golang endpoint"
	@echo "  make help            		- Show this help message"

# Start all services
up:
	docker compose up -d --build

# Stop all services
down:
	docker compose down

# Restart all services
restart:
	docker compose restart

# View logs from all services
logs:
	docker compose logs -f

# View logs from specific services
logs-alloy:
	docker compose logs -f alloy

# View logs from Loki service
logs-loki:
	docker compose logs -f loki

# View logs from Grafana service
logs-grafana:
	docker compose logs -f grafana

# Reload Alloy service
alloy-reload:
	curl -X POST http://localhost:12345/-/reload

# Clean up everything
clean:
	docker compose down -v
	rm -rf ./docker/data/*

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

# Generate test load
test-load:
	@echo "Generating load on application endpoints..."
	@for i in {1..5}; do \
		curl -s "http://localhost:18080/demo/work" > /dev/null & \
		curl -s "http://localhost:18080/demo/cpu" > /dev/null & \
		curl -s "http://localhost:18080/demo/memory" > /dev/null; \
		sleep 1; \
	done
	@echo "Load test completed."
