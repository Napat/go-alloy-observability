# Golang + OpenTelemetry + Zap + Alloy

- [หลักการ แนวคิด และวิธีการใช้งาน Grafana Loki](docs/Loki.md)


## Diagram การทำงาน

```mermaid
flowchart TB
    %% Define components with ports
    Client["💻 Web/Mobile Clients"]
    GoApp["🚀 Go Application<br/>(REST API :18080, pprof :18080/debug/pprof)"]
    Alloy["📡 Alloy<br/>OTLP<br/>gRPC:4317,HTTP:4318"]
    Loki["📝 Loki<br/>(:3100)"]
    Mimir["📊 Mimir<br/>(:9009)"]
    Tempo["🔍 Tempo<br/>(:3200)"]
    Pyroscope["📈 Pyroscope<br/>(:4040)"]
    Grafana["🎯 Grafana<br/>(:13000)"]
    DevOps(["👤 DevOps/SRE"])

    %% Client to Application flow
    Client -->|"REST API :18080"| GoApp
    
    %% Telemetry flow from Go Application to Alloy
    GoApp -->|"OTLP HTTP Push :4318<br/>logs, metrics, traces"| Alloy

    %% Data flow from Alloy to storage systems
    Alloy -->|"HTTP Push :3100<br/>logs"| Loki
    Alloy -->|"HTTP Push :9009<br/>metrics"| Mimir
    Alloy -->|"HTTP Push :3200<br/>traces"| Tempo
    Alloy -->|"HTTP Push :4040<br/>profiles"| Pyroscope
    
    %% Profile pull from Alloy to Go Application
    Alloy -.->|"HTTP Pull :18080/debug/pprof<br/>pprof profiles"| GoApp

    %% Storage to Grafana
    Loki -->|"HTTP Pull Query :3100"| Grafana
    Mimir -->|"HTTP Pull Query :9009"| Grafana
    Tempo -->|"HTTP Pull Query :3200"| Grafana
    Pyroscope -->|"HTTP Pull Query :4040"| Grafana
    
    %% DevOps at bottom
    Grafana -->|"Monitor & Analyze :13000"| DevOps

    %% Modern styling (unchanged)
    classDef default fill:#ffffff,stroke:#2d3748,stroke-width:2px;
    classDef client fill:#f43f5e,color:#ffffff,stroke:#e11d48,stroke-width:2px;
    classDef app fill:#7c3aed,color:#ffffff,stroke:#6d28d9,stroke-width:2px;
    classDef collector fill:#06b6d4,color:#ffffff,stroke:#0891b2,stroke-width:2px;
    classDef storage fill:#3b82f6,color:#ffffff,stroke:#2563eb,stroke-width:2px;
    classDef viz fill:#10b981,color:#ffffff,stroke:#059669,stroke-width:2px;
    classDef user fill:#6366f1,color:#ffffff,stroke:#4f46e5,stroke-width:2px;
    
    class Client client;
    class GoApp app;
    class Alloy collector;
    class Loki,Mimir,Tempo,Pyroscope storage;
    class Grafana viz;
    class DevOps user;
```

## ทดสอบส่ง Log ผ่าน OTLP ด้วย curl ไปยัง Alloy

Flow: `curl -> Alloy -> Loki <- Grafana Query`

```sh
# curl OLTP log to Alloy
make test-alloy-logs
```

ตรวจสอบใน Grafana

- เปิด Grafana ที่ http://localhost:13000
- ไปที่ Explore -> เลือก Data source เป็น Loki
- ลองใช้ LogQL query ดูข้อมูล

```LogQL
{service_name="test-service"}`
```

หรือ

```LogQL
{service_name=~".+"} | json
```

กดดู Query inspector -> Data ควรจะเห็นข้อมูลที่ส่งเข้ามา

## Golang + OpenTelemetry + Zap + Alloy

Alloy เป็น distribution ของ OpenTelemetry Collector ซึ่งเป็นเครื่องมือสำหรับการเก็บรวบรวม ประมวลผล และส่งออก telemetry data เช่น logs, traces, และ metrics โดยเฉพาะอย่างยิ่งสำหรับการส่ง log ไปยัง Alloy ผ่าน OTLP ด้วย HTTP ผู้ใช้สามารถใช้ Golang ในการสร้าง application ที่ generate log โดยใช้ Zap ซึ่งเป็น logging library ที่มีประสิทธิภาพสูง และส่ง log เหล่านั้นไปยัง Alloy เพื่อการวิเคราะห์ต่อไป

OTLP (OpenTelemetry Protocol) เป็น protocol ที่ออกแบบมาเพื่อส่ง telemetry data ไปยัง backend ที่รองรับ เช่น Alloy โดยสามารถใช้ผ่าน HTTP หรือ gRPC สำหรับกรณีโปรเจคนี้ จะใช้ HTTP ซึ่งจะต้องทำการส่งข้อมูลผ่าน HTTP POST ไปยัง endpoint ที่กำหนด เช่น `http://localhost:4318/v1/logs`

Zap เป็น logging library สำหรับ Golang ที่มีโครงสร้างและประสิทธิภาพสูง แต่โดยปกติแล้ว Zap ไม่รองรับ OTLP โดยตรง ดังนั้น การผสานรวมกับ OTLP จึงต้องใช้ bridge หรือ plugin ที่ช่วยแปลง log record จาก Zap เป็น format ของ OpenTelemetry เพื่อให้สามารถส่งผ่าน OTLP Protocol (`Emit`) ไปยัง Alloy ได้ ดูตัวอย่างได้ที่ `pkg/otlp/otlp.go`

```sh
# Generate random log via golang endpoint
curl http://localhost:18080/demo/logs

# Generate random logs via golang endpoint
make test-go-random-logs
```


Generate some load on the application:

```sh
for i in {1..5}; do
  curl -s "http://localhost:18080/demo/work" > /dev/null &
  curl -s "http://localhost:18080/demo/cpu" > /dev/null &
  curl -s "http://localhost:18080/demo/memory" > /dev/null
  sleep 1
done
```






## 


## xxx

```sh
docker compose down
docker compose up --force-recreate -d
```


## Reference

- [How-to-ingest-logs-with-alloy-or-the-opentelemetry-collector](https://grafana.com/blog/2025/02/24/grafana-loki-101-how-to-ingest-logs-with-alloy-or-the-opentelemetry-collector/)
- [Scaling Observability to 50TB+ of Telemetry a Day at Wise](https://www.youtube.com/watch?v=Sd8epoCHoi0)
