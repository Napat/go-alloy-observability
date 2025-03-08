# Mimir and OpenTelemetry Metrics

Mimir เป็น time series database ที่ใช้เก็บ Metrics โดยเฉพาะ โดยสามารถรับข้อมูลเข้ามาในรูปแบบ Prometheus format และเก็บไว้ใน object storage หรืออาจจะตั้งค่าให้เก็บใน local file system ก็ได้

GoApp -> Alloy -> Mimir

Alloy จะรับข้อมูล Metrics เข้าไปในมาตรฐาน OpenTelemetry Metrics
จากนั้นจะแปลงให้อยู่ในรูปแบบของ Permetheus Metrics แล้วจึงส่งต่อไปให้ Mimir เพื่อเก็บข้อมูลเอาไว้

## OpenTelemetry Metrics

การเก็บ Metrics ใน Go Application ด้วย OpenTelemetry มีประเภทของ metrics พื้นฐาน 4 แบบ

1. Counter

- ใช้นับจำนวนที่เพิ่มขึ้นเรื่อยๆ เช่น จำนวน requests, errors
- ค่าจะเพิ่มขึ้นเท่านั้น ไม่มีการลดลง

```golang
// สร้าง Counter ใหม่ โดยใช้ชื่อ "http_requests_total"
requestCounter, _ := meter.Int64Counter(
    "http_requests_total",
    metric.WithDescription("Total number of HTTP requests"),
)

// การใช้งาน Counter โดยสั่งให้ค่าเพิ่มค่าขึ้น +1 ทุกครั้งที่มีการเรียกใช้งาน 
requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
```

2. UpDown Counter

- คล้าย Counter แต่ค่าสามารถลดลงได้
- ใช้กับค่าที่มีการเพิ่ม/ลด เช่น จำนวน active connections

```golang
activeConnections, _ := meter.Int64UpDownCounter(
    "active_connections",
    metric.WithDescription("Number of active connections"),
)

// การใช้งาน
activeConnections.Add(ctx, 1)  // เพิ่มค่า
activeConnections.Add(ctx, -1) // ลดค่า
```

3. Gauge

- ค่าที่สามารถขึ้นลงได้ เช่น memory usage, CPU usage
- เหมาะกับค่าที่เปลี่ยนแปลงตลอดเวลา

```golang
// สร้าง Gauge ใหม่ โดยใช้ชื่อ "memory_usage_bytes"
memoryGauge, _ := meter.Float64ObservableGauge(
    "memory_usage_bytes",
    metric.WithDescription("Current memory usage"),
)

// การใช้งาน Gauge โดยสั่งให้ค่าเปลี่ยนแปลงตลอดเวลา โดยใช้ค่า memory จาก runtime.MemStats.Alloc 
meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
    o.ObserveFloat64(memoryGauge, float64(runtime.MemStats.Alloc))
    return nil
})
```

4. Histogram

- ใช้วัดการกระจายของข้อมูล เช่น request duration, response size
- แสดงผลเป็น buckets ของข้อมูล

```golang
latencyHistogram, _ := meter.Float64Histogram(
    "http_request_duration_seconds",
    metric.WithDescription("HTTP request duration"),
)

// การใช้งาน
start := time.Now()
// ... do something ...
duration := time.Since(start).Seconds()
latencyHistogram.Record(ctx, duration, metric.WithAttributes(attrs...))
```

## Mimir

Mimir จะรับข้อมูลเข้ามาในรูปแบบ Prometheus format 
โดยปกติแล้วเรามักจะออกแบบให้ไปเก็บข้อมูลไว้ใน object storage เช่น AWS S3, GCS, Azure Blob Storage, MinIO(S3 onprimise), etc.
การเก็บข้อมูล(long-term data storage) ใน object storage จะดีกว่าการเก็บใน local storage เมื่อนำไปใช้ในระบบที่ทำ AZ(Availability Zone) 
เนื่องจากประหยัดค่าใช้จ่ายโดยเฉพาะเรื่อง Network cost(sync data ให้แต่ละ AZ ต้องเท่ากัน), ดูแลง่ายกว่าด้วย
(object storage allowing us to take advantage of this ubiquitous, cost-effective, high-durability technology)

Prometheus มีแค่ Counter และ Gauge เป็นหลัก (ไม่เหมือน OpenTelemetry Metrics) 

ดังนั้นเมื่อส่ง UpDown Counter (เป็น metric type ใน OpenTelemetry) ไปยัง Alloy และแปลงเป็น Prometheus format ผ่าน otelcol.exporter.prometheus มันจะกลายเป็น metric แบบ gauge ใน Mimir

## Reference

- [Mimir](https://grafana.com/docs/mimir/latest/get-started/play-with-grafana-mimir/#download-tutorial-configuration)