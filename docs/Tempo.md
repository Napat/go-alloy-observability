# Tempo and Distributed Tracing

GoApp -> Alloy -> Tempo

Tempo เป็น distributed tracing backend ที่มีประสิทธิภาพสูง รองรับการจัดเก็บ traces ในปริมาณมาก และสามารถทำงานร่วมกับ OpenTelemetry ได้อย่างดี

## Distributed Tracing Concepts

### Trace
- Trace คือการบันทึกการทำงานของ request ตั้งแต่ต้นจนจบ
- ประกอบด้วย spans หลายๆ อันที่เชื่อมโยงกัน
- แต่ละ trace มี unique trace ID

### Span
- Span คือหน่วยย่อยที่สุดของ trace
- แสดงถึงการทำงานหนึ่งๆ เช่น HTTP request, database query
- ประกอบด้วย:
  - Name: ชื่อของ operation
  - Start/End time: เวลาเริ่มและจบการทำงาน
  - Attributes: ข้อมูลเพิ่มเติม เช่น HTTP method, status code
  - Events: เหตุการณ์สำคัญระหว่างการทำงาน
  - Parent span: span ที่เป็นต้นทางของ span นี้

## การใช้งาน OpenTelemetry Tracing ใน Go

### 1. การสร้าง Tracer

```go
import "go.opentelemetry.io/otel"

tracer := otel.Tracer("service-name")
```

### 2. การสร้าง Span

```go
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

// เพิ่ม attributes
span.SetAttributes(
    attribute.String("http.method", "GET"),
    attribute.Int("http.status_code", 200),
)

// บันทึก events
span.AddEvent("starting-operation")
```

### 3. การเชื่อมโยง Spans

```go
// Parent span
ctx, parentSpan := tracer.Start(ctx, "parent-operation")
defer parentSpan.End()

// Child span (จะเชื่อมโยงกับ parent โดยอัตโนมัติ)
ctx, childSpan := tracer.Start(ctx, "child-operation")
defer childSpan.End()
```

### การ Query Traces ใน Grafana

Basic Queries

```txt
# ค้นหาด้วย Trace ID
{traceID="abcd1234..."}

# ค้นหาตาม Service
{service_name="my-service"}

# ค้นหา Error Traces
{status="error"}
```

Advanced Queries

```txt
# ค้นหา Traces ที่มี Duration สูง
{service_name="my-service"} | duration > 1s

# ค้นหาตาม HTTP Status
{service_name="my-service"} | http.status_code=500
```

### ## Best Practices
1. Sampling

   - ใช้ sampling strategy ที่เหมาะสมเพื่อควบคุมปริมาณ traces
   - เก็บ traces ทั้งหมดสำหรับ errors
   - ใช้ probability sampling สำหรับ successful requests

2. Span Attributes
   
   - ตั้งชื่อ spans ให้มีความหมาย
   - ใส่ attributes ที่จำเป็นต่อการ debug
   - อย่าใส่อะไรที่ไม่จำเป็นเข้าไปใน attributes เพราะมีผลต่อ performance
   - หลีกเลี่ยงการใส่ข้อมูลที่ sensitive

3. Context Propagation
   
   - ส่งต่อ context ระหว่าง services อย่างถูกต้อง
   - ใช้ standard headers สำหรับ trace propagation
