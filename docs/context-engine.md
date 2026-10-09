# Backend Context Engine — Technical Documentation

## 1. Architecture Overview

Backend Context Engine dirancang untuk aplikasi pembelajaran realtime yang menerima stream transkrip teks dari Speech-to-Text (STT) frontend secara bertahap (*chunks/deltas*), kemudian mengekstrak konteks dan konsep pembelajaran secara *asynchronous* melalui AI background worker, lalu mengirimkan hasilnya kembali ke frontend melalui koneksi WebSocket.

```mermaid
graph TD
    FE["Frontend (Client)"]
    API["Gin HTTP API Server"]
    DB[("PostgreSQL Database")]
    Worker["Background Context Worker"]
    AI["ContextEngine (Mock / LLM)"]
    WSHub["WebSocket Hub"]

    FE -- "1. POST /sessions/:id/transcripts" --> API
    API -- "2. Save chunk (status=pending)" --> DB
    API -- "3. 202 Accepted" --> FE

    Worker -- "4. Poll pending chunks" --> DB
    Worker -- "5. Analyze (Recent + Current Context)" --> AI
    AI -- "6. Context result" --> Worker
    Worker -- "7. Save Context & status=completed" --> DB
    Worker -- "8. Publish Event" --> WSHub
    WSHub -- "9. WS context_update" --> FE
```

Prinsip arsitektur utama:
- **Fast Ingestion**: Endpoint transkrip segera merespons `202 Accepted` tanpa menunggu proses AI.
- **Asynchronous Processing**: Background worker memproses antrean transkrip dari database.
- **Decoupled Realtime Layer**: Domain service mempublikasikan event melalui abstraksi `EventPublisher`, tidak terikat langsung pada WebSocket.
- **Pluggable AI Engine**: Abstraksi `ContextEngine` memudahkan pertukaran engine (Mock, Gemini, OpenAI, Claude) tanpa mengubah business logic.

---

## 2. Request Flow

```mermaid
sequenceDiagram
    autonumber
    participant FE as Frontend
    participant API as Ingestion API
    participant DB as PostgreSQL
    participant Worker as Context Worker
    participant AI as Context Engine
    participant WS as WebSocket Hub

    Note over FE,WS: Setup Sesi Pembelajaran
    FE->>API: POST /api/v1/sessions
    API->>DB: Simpan Session (active)
    API-->>FE: 201 Created (sessionId)
    FE->>WS: Connect /ws/sessions/:sessionId
    WS-->>FE: 101 Switching Protocols

    Note over FE,WS: Ingestion & Pemrosesan Transkrip
    FE->>API: POST /api/v1/sessions/:sessionId/transcripts (sequence: 1, text: "...")
    API->>DB: Simpan Transcript (status=pending, unique sequence)
    API-->>FE: 202 Accepted (accepted: true, sequence: 1)

    Worker->>DB: Query pending chunks
    Worker->>DB: Update status = processing
    Worker->>DB: Fetch session metadata & recent chunks (N=5)
    Worker->>AI: Analyze(ContextInput)
    AI-->>Worker: ContextResult (concept_update / formula / no_change)
    alt Ada Perubahan Konteks
        Worker->>DB: Simpan ContextChunk
        Worker->>WS: Broadcast Event (context_update)
        WS-->>FE: WebSocket Frame: context_update
    end
    Worker->>DB: Update status = completed
```

---

## 3. Database Schema

### 3.1 `sessions`
| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | UUID (PK) | Auto-generated UUID |
| `subject` | VARCHAR(100) | Mata pelajaran (misal: "Fisika") |
| `topic` | VARCHAR(100) | Topik bahasan (misal: "Dinamika") |
| `subtopic` | VARCHAR(100) | Sub-topik (misal: "Hukum Newton") |
| `status` | VARCHAR(20) | `active`, `completed`, `cancelled` |
| `created_at` | TIMESTAMPTZ | Waktu pembuatan |
| `updated_at` | TIMESTAMPTZ | Waktu perubahan |

### 3.2 `transcript_chunks`
| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | UUID (PK) | Auto-generated UUID |
| `session_id` | UUID (FK/Index) | ID Sesi terkait |
| `sequence` | BIGINT | Urutan kronologis chunk ($1, 2, 3, \dots$) |
| `text` | TEXT | Konten teks ucapan |
| `status` | VARCHAR(20) | `pending`, `processing`, `completed`, `failed` |
| `created_at` | TIMESTAMPTZ | Waktu diterima |
| `updated_at` | TIMESTAMPTZ | Waktu perubahan |

> **Constraint Penting**: Composite unique index pada `(session_id, sequence)` untuk mencegah duplicate sequence.

### 3.3 `context_chunks`
| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | UUID (PK) | Auto-generated UUID |
| `session_id` | UUID (FK/Index) | ID Sesi terkait |
| `sequence` | BIGINT | Sequence transkrip yang memicu context ini |
| `type` | VARCHAR(50) | `concept_update`, `formula_detected`, dll. |
| `data` | JSONB | Data konteks terstruktur (summary, formula, keywords) |
| `created_at` | TIMESTAMPTZ | Waktu context dibentuk |
| `updated_at` | TIMESTAMPTZ | Waktu perubahan |

---

## 4. API Endpoints

Format respon standar:
```json
{
  "meta": {
    "success": true,
    "message": "Pesan deskriptif"
  },
  "data": {},
  "errors": null
}
```

### 4.1 Session Management
- **`POST /api/v1/sessions`**
  - Status: `201 Created`
  - Body:
    ```json
    {
      "subject": "Fisika",
      "topic": "Dinamika",
      "subtopic": "Hukum Newton"
    }
    ```
- **`GET /api/v1/sessions/:id`**
  - Status: `200 OK` (atau `404 Not Found`)

### 4.2 Transcript Ingestion
- **`POST /api/v1/sessions/:sessionId/transcripts`**
  - Status: `202 Accepted`
  - Body:
    ```json
    {
      "sequence": 1,
      "text": "Hari ini kita akan membahas Hukum Newton."
    }
    ```
  - Response:
    ```json
    {
      "meta": { "success": true, "message": "Transcript accepted" },
      "data": { "accepted": true, "sequence": 1 }
    }
    ```
  - Error:
    - `404 Not Found`: Jika session ID tidak terdaftar.
    - `409 Conflict`: Jika sequence sudah pernah diterima untuk sesi tersebut.
    - `422 Unprocessable Entity`: Jika payload tidak valid (sequence $< 1$, text kosong).

- **`GET /api/v1/sessions/:sessionId/transcripts`**
  - Status: `200 OK`
  - Mengembalikan list chunk yang diurutkan berdasarkan `sequence ASC`.

### 4.3 Context History
- **`GET /api/v1/sessions/:sessionId/contexts`**
  - Status: `200 OK`
  - Mengembalikan daftar context chunks terstruktur yang telah dihasilkan AI.

---

## 5. WebSocket Protocol

### 5.1 Endpoint
```
WS /ws/sessions/:sessionId
```

### 5.2 Format Payload Event
Ketika AI menghasilkan context update:
```json
{
  "event": "context_update",
  "sessionId": "4a7f9201-1b2c-...",
  "sequence": 4,
  "timestamp": "2026-10-09T10:30:00Z",
  "data": {
    "type": "formula_detected",
    "concept": "Hukum II Newton",
    "summary": "Hubungan percepatan berbanding lurus dengan gaya dan berbanding terbalik dengan massa.",
    "keywords": ["gaya", "massa", "percepatan", "hukum newton"],
    "formula": "F = m × a"
  }
}
```

---

## 6. Background Worker

Worker (`ContextWorker`) berjalan dalam background goroutine independen:
1. **Polling interval**: Diatur melalui env `CONTEXT_WORKER_INTERVAL` (default: `1s`).
2. **Duplicate Protection**: Transkrip yang diambil langsung diubah statusnya menjadi `processing`.
3. **Context Sliding Window**: Worker mengambil $N$ transkrip terakhir (`CONTEXT_RECENT_CHUNKS`, default: `5`) up to current sequence, ditambah *Current Context* terakhir.
4. **Resilience**: Jika pemrosesan sebuah chunk gagal, status ditandai `failed` dan worker tetap melanjutkan pemrosesan chunk lain tanpa crash.
5. **Graceful Shutdown**: Menerima `context.Context` signal pembatalan sehingga tidak memutus operasi database di tengah jalan.

---

## 7. Context Engine Abstraction

Engine didefinisikan melalui interface sederhana:
```go
type ContextEngine interface {
    Analyze(ctx context.Context, input ContextInput) (*ContextResult, error)
}
```

- **`MockContextEngine`**: Implementasi deterministik berbasis pola kata kunci untuk pengujian tanpa API key.
- Mendukung evaluasi semantic: jika tidak ada konsep esensial yang berubah, engine menghasilkan `type: "no_change"` sehingga sistem tidak mengirimkan broadcast spam ke frontend (§23).

---

## 8. Error Handling

Mengikuti hierarki domain error dan HTTP status code:
- `400 Bad Request`: Payload JSON cacat.
- `404 Not Found`: Entity (Session) tidak ditemukan.
- `409 Conflict`: Sequence transkrip duplikat.
- `422 Unprocessable Entity`: Validasi field gagal (`sequence < 1`, `text` kosong).
- `500 Internal Server Error`: Kesalahan infrastruktur database atau server internal.

---

## 9. Configuration

Konfigurasi dibaca dari environment variables atau file `.env`:

| Variable | Default | Deskripsi |
|---|---|---|
| `APP_ENV` | `development` | Environment (`development` / `production`) |
| `APP_PORT` | `8080` | Port HTTP API server |
| `DATABASE_URL` | - | PostgreSQL Connection String |
| `CONTEXT_RECENT_CHUNKS` | `5` | Jumlah transkrip terakhir sebagai konteks AI |
| `CONTEXT_WORKER_INTERVAL` | `1s` | Interval polling background worker |
| `AI_PROVIDER` | `mock` | Provider engine (`mock`, `gemini`, dll.) |
| `AI_API_KEY` | - | API key untuk AI provider |
| `WEBSOCKET_MAX_CONNECTIONS` | `100` | Batas maksimum client WebSocket aktif |

---

## 10. Local Development & Menjalankan Aplikasi

1. Buat database PostgreSQL lokal (misal: `joints_be`).
2. Pastikan file `.env` sudah diatur:
   ```env
   DATABASE_URL=postgres://postgres:password@localhost:5432/joints_be?sslmode=disable
   ```
3. Jalankan server:
   ```bash
   go run cmd/main.go
   ```
4. Server akan otomatis melakukan auto-migrate tabel pada mode `development` dan memulai worker & WebSocket hub.

---

## 11. Testing

Jalankan seluruh rangkaian unit & integration tests:
```bash
go test -v ./...
```

Cakupan pengujian:
- **`modules/session/service`**: Create dan retrieval sesi.
- **`modules/transcript/service`**: Ingestion, 202 Accepted, duplicate sequence rejection (409), session validation.
- **`modules/context/service`**: Context generation, semantic `no_change` filtering, AI provider failure handling.
- **`pkg/websocket`**: Registration, isolation per session, broadcast delivery, connection limit.
- **`modules/context/tests`**: End-to-end integration test dari ingestion 5 chunks hingga broadcast event dan update konteks.

---

## 12. Future RAG Architecture

Arsitektur saat ini telah dirancang untuk siap menerima RAG (*Retrieval-Augmented Generation*) tanpa perombakan struktur:

```mermaid
graph LR
    Worker["Context Worker"]
    Engine["ContextEngine"]
    Retriever["Retriever Interface"]
    VectorDB[("Vector Database / pgvector")]
    LLM["LLM Provider (Gemini / OpenAI)"]

    Worker --> Engine
    Engine --> Retriever
    Retriever -- "Embed & Vector Search" --> VectorDB
    VectorDB -- "Relevant Material Docs" --> Retriever
    Retriever --> Engine
    Engine -- "Prompt + Retrieved Docs + Context" --> LLM
    LLM --> Engine
    Engine --> Worker
```

Ketika modul RAG diaktifkan, cukup menambahkan implementasi `Retriever` ke provider `ContextEngine` tanpa perlu mengubah Controller, Repository, maupun WebSocket Layer.

