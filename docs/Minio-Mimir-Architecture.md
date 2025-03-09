# MinIO and Mimir Architecture

```mermaid
graph TB
    %% Define Application Components
    App["Application"]:::app --> Alloy["Alloy"]:::collector
    
    %% Define Mimir Components
    subgraph "Mimir Cluster"
        Mimir1["Mimir-1"]:::mimir
        Mimir2["Mimir-2"]:::mimir
        Mimir3["Mimir-3"]:::mimir
    end
    
    %% Define MinIO Storage
    subgraph "Object Storage"
        MinIO["MinIO"]:::storage
        MinIOBuckets["Buckets:<br/>- mimir<br/>- mimir-rules"]:::storage
    end
    
    %% Define Load Balancer and UI Access
    LoadBalancer["Load Balancer"]:::lb
    Grafana["Grafana"]:::ui
    DevOps["DevOps/SRE"]:::user
    
    %% Define Data Flow
    Alloy -->|"Prometheus Format"| LoadBalancer
    LoadBalancer -->|"Write/Query"| Mimir1
    LoadBalancer -->|"Write/Query"| Mimir2
    LoadBalancer -->|"Write/Query"| Mimir3
    
    %% MinIO Connections
    Mimir1 -->|"Store/Retrieve"| MinIO
    Mimir2 -->|"Store/Retrieve"| MinIO
    Mimir3 -->|"Store/Retrieve"| MinIO
    MinIO --- MinIOBuckets
    
    %% UI Access
    LoadBalancer -->|"Query(HTTP Pull):19009"| Grafana
    MinIO -->|":9001<br/>MinIO Web UI"| LoadBalancer
    LoadBalancer -->|":19001<br/>MinIO Web UI"| DevOps
    Grafana -->|"View Metrics"| DevOps
    
    %% Styling
    classDef default fill:#ffffff,stroke:#2d3748,stroke-width:2px
    classDef mimir fill:#7c3aed,color:#ffffff,stroke:#6d28d9,stroke-width:2px
    classDef storage fill:#3b82f6,color:#ffffff,stroke:#2563eb,stroke-width:2px
    classDef lb fill:#06b6d4,color:#ffffff,stroke:#0891b2,stroke-width:2px
    classDef ui fill:#10b981,color:#ffffff,stroke:#059669,stroke-width:2px
    classDef user fill:#f43f5e,color:#ffffff,stroke:#e11d48,stroke-width:2px
    classDef app fill:#f97316,color:#ffffff,stroke:#ea580c,stroke-width:2px
    classDef collector fill:#4b5cf6,color:#ffffff,stroke:#4c3aed,stroke-width:2px
```

## สถาปัตยกรรมและการทำงาน

### 1. การไหลของข้อมูล (Data Flow)

- **Application → Alloy**:
  - แอปพลิเคชันส่ง metrics ไปยัง Alloy ผ่าน OTLP HTTP Push
  - Alloy ทำหน้าที่รวบรวมและแปลงข้อมูลให้อยู่ในรูปแบบที่ Mimir เข้าใจได้

- **Alloy → Mimir**:
  - Alloy แปลงข้อมูลให้อยู่ในรูปแบบ Prometheus remote write
  - Load Balancer กระจาย requests ไปยัง Mimir instances ต่างๆ เพื่อการทำงานแบบ high availability

### 2. Mimir Cluster

- **High Availability**:
  - ประกอบด้วย Mimir 3 instances ทำงานพร้อมกัน
  - แต่ละ instance สามารถรับ write และ query requests ได้
  - หากมี instance ใดล้มเหลว instances อื่นยังคงให้บริการได้ต่อเนื่อง

- **การจัดการข้อมูล**:
  - ทุก instance มีสิทธิ์เข้าถึง MinIO เท่าเทียมกัน
  - ใช้ S3 compatible API ในการอ่าน/เขียนข้อมูลกับ MinIO
  - รองรับการทำ horizontal scaling โดยเพิ่ม instances ได้ตามต้องการ

### 3. MinIO Object Storage

- **การจัดเก็บข้อมูล**:
  - ทำหน้าที่เป็น distributed object storage
  - รองรับการเก็บข้อมูลขนาดใหญ่และการเข้าถึงแบบ concurrent
  - มีระบบ data replication ในตัวเพื่อความทนทานของข้อมูล
  - ลดความซับซ้อนในการ sync ข้อมูลระหว่าง availability zones
  - ประหยัดค่าใช้จ่ายด้านเครือข่ายเมื่อเทียบกับระบบไฟล์แบบดั้งเดิม

- **ความสามารถหลัก**:
  - รองรับ S3 API ทำให้ง่ายต่อการพัฒนาและบำรุงรักษา
  - มีระบบรักษาความสอดคล้องของข้อมูล (consistency)
  - รองรับการทำ disaster recovery
  - สามารถขยายระบบได้ตามปริมาณข้อมูลที่เพิ่มขึ้น

- **Buckets**:
  - `mimir`: เก็บข้อมูล metrics ทั้งหมด
    - Time series data
    - Metadata
    - Indexes
  - `mimir-rules`: เก็บ configuration สำหรับ alerting rules
    - Alert rules
    - Recording rules
    - Alert templates

### 4. การเข้าถึงระบบ

- **Metrics Access**:
  - Load Balancer เปิด port 19009 สำหรับ Prometheus API
  - Grafana ใช้ port นี้ในการ query ข้อมูลจาก Mimir
  - รองรับ PromQL queries และ Prometheus remote write

- **Management**:
  - MinIO Web UI: เข้าถึงผ่าน port 19001
  - ใช้สำหรับ monitor storage usage และจัดการ buckets
  - มี dashboard แสดงสถานะและ performance metrics

- **Monitoring**:
  - Grafana dashboards แสดงข้อมูล metrics จาก Mimir
  - สามารถสร้าง custom dashboards และ alerts
  - รองรับการ visualize ข้อมูลในหลากหลายรูปแบบ