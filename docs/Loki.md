# แนวคิดหลักในการใช้งาน Grafana Loki

Grafana Loki เป็นระบบ Logging แบบ Open-Source ที่พัฒนาโดย Grafana Labs ออกแบบมาเพื่อจัดเก็บและค้นหา Logs จำนวนมากได้อย่างมีประสิทธิภาพ โดยชูจุดเด่นคือความเรียบง่ายและต้นทุนต่ำเมื่อเทียบกับระบบ Logging อื่น ๆ เช่น Elasticsearch 

Loki ถูกสร้างมาเพื่อทำงานร่วมกับ Grafana stack อื่นๆและ Prometheus ได้อย่างลงตัว ช่วยให้ทีม DevOps และ Software Engineers สามารถเชื่อมโยง Logs ได้ง่ายผ่านการใช้ Labels และภาษาคิวรีอย่าง LogQL

Loki ใช้แนวคิดที่เรียกว่า "Indexing Metadata Only" ซึ่งหมายความว่ามันจะทำการ Index เฉพาะ Metadata (เช่น Labels) ของ Log Entry เท่านั้น โดยไม่ Index ข้อมูลเนื้อหาภายใน Log (Log Message) เอง 

Loki ไม่เหมือน Elasticsearch(ES) ใน ELK ที่ทำ pre-indexing ทุกๆอย่าง ทำให้ cost ของการใช้ Loki ถูกกว่ามากๆแลกมากับ speed ในการ queries ข้อมูลที่จะช้ากว่า ES แต่ก็มีการชดเชยด้วย alogorithm ที่ช่วยให้เราสามารถ query ข้อมูลได้เร็วขึ้นเช่น ใช้ bloom filters in search เป็นต้น

Bloom Filters ใน Loki ช่วยให้ค้นหา Logs ได้เร็วขึ้นโดยไม่ต้องสร้าง Index เต็มรูปแบบ โดยทำหน้าที่เป็นด่านกรองเบื้องต้น ว่า Chunk ไหนควรค้นหา Chunk ไหนควรข้าม ช่วยลดภาระและประหยัดทรัพยากรไปได้เยอะ

✅ ข้อดีของการไม่ใช้ Indexing เต็มรูปแบบ

1.	ประหยัดพื้นที่จัดเก็บ: เนื่องจากไม่ต้องสร้างดัชนีขนาดใหญ่
2.	ต้นทุนต่ำ: สามารถจัดเก็บ Log ได้มากขึ้นในราคาเดียวกัน
3.	การรวมกับ Prometheus ได้ดี: Loki ใช้ Labels แบบเดียวกับ Prometheus ทำให้ใช้ Query แบบเดียวกัน (LogQL) ได้ง่าย
4.	การติดตั้งและบำรุงรักษาง่าย: ไม่ต้องจัดการดัชนีหรือทำ Index Rebuilding

❌ ข้อเสียและข้อจำกัด

1.	ประสิทธิภาพการค้นหาช้า: โดยเฉพาะเมื่อค้นหาข้อความแบบ Full-Text ใน Log Message
2.	ใช้หน่วยความจำมากขึ้นในการค้นหา: การค้นหา In-Stream ต้องโหลดข้อมูล Log มาประมวลผลใน Memory มากกว่าระบบที่มี Index
3.	ไม่เหมาะกับการ Query ที่ซับซ้อน: เช่น การค้นหาด้วยเงื่อนไข Regular Expression (regex) ขนาดใหญ่

📦 ตัวอย่างการใช้งาน LogQL

```logql
# ค้นหา Log ของแอป nginx ที่มีข้อความ "error" อยู่
{app="nginx"} |= "error"

# ค้นหาจาก namespace production และใช้ regex เพื่อกรอง Log
{namespace="prod"} |~ "timeout|failed"

# คำนวณอัตราของ Logs ที่มีคำว่า error ใน 5 นาที
rate({app="nginx"} |= "error" [5m])

# การใช้ Regex เพื่อค้นหา Logs ที่มีทั้ง error และ timeout ในลำดับใดก็ได้
{namespace="prod"} |~ "error.*timeout"
```

🔥 Loki เหมาะกับใคร?

•	ระบบที่มีการจัดเก็บ Log จำนวนมากแต่ต้องการประหยัดพื้นที่
•	DevOps ที่ใช้ Prometheus อยู่แล้วและต้องการระบบ Logging ที่เข้ากันได้
•	ใช้กับ Kubernetes เพราะ Labels แบบเดียวกันช่วยให้การเชื่อมโยง Logs กับ Metrics ทำได้ง่าย

## 🧠 หลักการออกแบบ Metadata Labels เพื่อใช้ใน Loki

1.	เลือก Labels ที่มี Cardinality ต่ำ
    Cardinality หมายถึงจำนวนค่าที่เป็นไปได้ของ Label นั้น ๆ เช่น
        •	✅ env (dev, staging, prod) => Cardinality ต่ำ (3 ค่า)
        •	❌ pod_name (nginx-1, nginx-2, … , nginx-1000) => Cardinality สูง (1000 pods)
        เหตุผล: Cardinality สูงเนื่องจาก pod_name เป็น Label ใน Kubernetes Cluster ที่มี 1,000 Pods จะสร้าง Index จำนวนมาก ทำให้การค้นหาช้าลงและใช้ Memory สูงขึ้น

2.	ใช้ Labels สำหรับการกรองข้อมูลหลัก (High-Level Filtering)
    •	✅ ตัวอย่างที่ดี: namespace, app, env, region
    •	❌ ตัวอย่างที่ไม่ควรใช้เป็น Label: timestamp, request_id, uuid (เพราะแต่ละ Log มีค่าที่ไม่ซ้ำกัน ทำให้ Cardinality สูง)

3.	ทำ Label ให้มีความหมายและใช้งานง่าย
    •	✅ ใช้ Label ที่สามารถอธิบายระบบได้ เช่น

    ```yaml
    {app="nginx", namespace="production", env="prod"}
    ```

    •   ❌ หลีกเลี่ยงการสร้าง Label แบบไร้โครงสร้าง เช่น

    ```yaml
    {random_key="value123"}  # ไม่รู้ว่าเอาไว้ใช้กรองอะไร
    ```

4.	ใช้ Labels สำหรับการจัดกลุ่มข้อมูล (Grouping) และหลีกเลี่ยงการใช้ Labels ซ้อนกันในการค้นหา (Nested Labels Search)
    •	✅ ตัวอย่างที่ดี:
    ```yaml
    {app="nginx", env="prod", region="us-east-1"}
    ```
    •	❌ ตัวอย่างที่ไม่ควรใช้เป็น Label:
    ```yaml
    {app="nginx", env="prod", region="us-east-1", timestamp="2023-01-01T00:00:00Z", response_time="150ms"}
    ```

## 🌼 Bloom Filter คืออะไร?

Bloom Filter เป็นโครงสร้างข้อมูลแบบ probabilistic data structure ที่ใช้ในการตรวจสอบว่า "ค่านี้อาจจะอยู่" หรือ "ค่านี้ไม่มีทางอยู่" ในชุดข้อมูล (Set) โดยใช้พื้นที่จัดเก็บน้อยและทำงานได้รวดเร็ว

ทำงานอย่างไร: ใช้การ Hash ค่า (เช่น คำหรือข้อความ) ด้วยหลายๆ ฟังก์ชัน Hash แล้วทำเครื่องหมายใน Bit Array
ผลลัพธ์:
•	ถ้าบอกว่า "ไม่มีทางอยู่" แปลว่าไม่มีจริงๆ (100% แน่นอน)
•	ถ้าบอกว่า "อาจจะอยู่" แปลว่าต้องไปเช็กข้อมูลจริงอีกครั้ง (เพราะมีโอกาสเกิด False Positive)

สมมติว่าเราต้องการค้นหาคำว่า error ใน Logs จำนวนมาก Bloom Filters จะช่วยบอกว่า Chunk ไหนที่อาจมีคำว่า error โดยไม่ต้องเปิดดูทุก Chunk

### 🛠️ Bloom Filter ช่วย Loki ได้อย่างไร?

Loki ใช้ Bloom Filters เพื่อช่วยกรองข้อมูล Log เบื้องต้นก่อนที่จะทำการค้นหาแบบ Full-Text ภายใน Log Stream โดยมีขั้นตอนดังนี้

1.	เก็บ Metadata ด้วย Bloom Filters: เมื่อ Logs ถูกเก็บเข้าระบบ Loki จะสร้าง Bloom Filters สำหรับแต่ละ Chunk ของ Log โดยกรองข้อมูลจาก Labels และคีย์เวิร์ดสำคัญ
2.	ค้นหาแบบเบื้องต้น: เมื่อผู้ใช้ Query Logs ผ่าน LogQL ระบบจะใช้ Bloom Filters เพื่อค้นหาว่า Chunk ไหนที่ "อาจจะมีข้อมูลที่ต้องการ"
3.	ค้นหาแบบละเอียด: Loki จะทำการอ่านเฉพาะ Chunks ที่ Bloom Filter บอกว่า "อาจจะมี" เท่านั้น แล้วทำการ Full-Text Search ภายใน Chunks เหล่านั้น
4.	ลดภาระระบบ: ช่วยลดจำนวน Chunks ที่ต้องอ่าน ทำให้การค้นหาเร็วขึ้น และไม่ต้องใช้หน่วยความจำสูงมาก

🎯 ตัวอย่างสถานการณ์: การค้นหา Log ด้วยข้อความ `error` ในแอป nginx

```logql
{app="nginx"} |= "error"
```

•	ระบบจะดูที่ Bloom Filters ของแต่ละ Chunk ว่ามีโอกาสพบคำว่า "error" ไหม
•	ถ้า Bloom Filter บอกว่า "ไม่มีทางอยู่" ก็ข้าม Chunk นั้นไปได้เลย
•	ถ้า Bloom Filter บอกว่า "อาจจะมี" ถึงจะไปทำการค้นหา Full-Text จริงๆ ใน Chunk นั้น

ข้อดีของ Bloom Filters

- ✅ ประหยัดพื้นที่:	Bloom Filter ใช้ Bit Array เล็กๆ จึงไม่เปลืองพื้นที่เก็บข้อมูล
- ✅ ค้นหาได้เร็วขึ้น:	ลดจำนวน Chunks ที่ต้องอ่าน ทำให้การ Query Logs เร็วขึ้น
- ✅ ไม่ต้องสร้างดัชนีขนาดใหญ่:	ช่วยให้ Loki เบาและประหยัดทรัพยากร

ข้อเสียของ Bloom Filters

- ❌ เกิด False Positive: บางครั้ง Bloom Filter อาจบอกว่า "อาจจะมี" แต่จริงๆ ไม่มีข้อมูลอยู่
- ❌ ลบข้อมูลไม่ได้: เมื่อเพิ่มข้อมูลเข้าไปใน Bloom Filter แล้ว จะลบออกไม่ได้ ทำให้ใช้ได้ดีกับข้อมูลที่ Append-only เท่านั้น 

**หากต้องการลบข้อมูลต้องทำการ Rebuild Bloom Filter ใหม่** ทำให้ต้องออกแบบการเก็บข้อมูลให้ดี

## 🧹 แล้ว Loki ลบ Logs ได้ยังไง?

🛠️ แม้ว่า Bloom Filter จะลบข้อมูลไม่ได้ แต่ใน Loki นั้นมีแนวทางแก้ไขข้อจำกัดของ Bloom Filterสามารถลบ Logs ได้ด้วยวิธีอื่น เช่น

1.	ใช้ Capped Chunks: Loki สามารถกำหนดขนาดหรือระยะเวลาของ Chunks ได้ เมื่อ Chunk เต็มหรือหมดอายุ สามารถลบออกทั้ง Chunk โดยไม่ต้องแก้ที่ Bloom Filter
2.	ใช้ "Counting Bloom Filter": เป็นเวอร์ชันพัฒนาของ Bloom Filter โดยแทนที่จะใช้ Bit Array แบบ 0/1 ก็เปลี่ยนเป็นตัวเลขที่นับจำนวนได้ ทำให้ลบข้อมูลได้โดยการลดค่าแทนการ Reset Bit
3.	แบ่งช่วงเวลา (Time-based Partitioning): จัดเก็บ Logs เป็นช่วงเวลา เช่น วันหรือชั่วโมง เมื่อต้องการลบ Logs เก่า ก็สามารถลบทั้ง Partition ได้โดยไม่กระทบกับ Bloom Filter

ใน Loki เราสามารถสั่งลบข้อมูลได้หลายวิธี เช่น

1.	Retention Policy: กำหนดระยะเวลาเก็บ Logs เช่น เก็บไว้ 30 วัน (`retention_period: 30d`) Logs เก่ากว่านั้นจะถูกลบอัตโนมัติ
2.	Compactor และ Chunks: Logs ใน Loki ถูกจัดเก็บใน Chunks ซึ่งจะมีการรวม(Compact) และลบ Chunks เก่าออกตามการตั้งค่าไว้
3.	Delete API: Loki มี API ให้ลบ Logs เฉพาะเจาะจงตามเงื่อนไข Labels ได้ เช่น

```sh
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"start":"2023-01-01T00:00:00Z","end":"2023-01-02T00:00:00Z","selectors":["{app=\"nginx\"}"]}' \
  http://localhost:3100/loki/api/v1/delete
```
