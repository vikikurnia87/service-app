# Planning: Backend Application — Echo + Bun + PostgreSQL + Redis

> **Tujuan dokumen ini:** Menjadi panduan arsitektur dan standar kode untuk membangun aplikasi backend yang clean, testable, dan scalable. Dokumen ini ditujukan untuk programmer atau AI agent yang akan mengimplementasikan detailnya.

---

## 1. Struktur Repository / Project

```
service-app/
├── cmd/
│   └── server/
│       └── main.go              # Entry point aplikasi
├── config/
│   └── config.go                # Load & parse konfigurasi (env / file)
├── internal/
│   ├── handler/                  # HTTP handler (controller layer)
│   │   └── user_handler.go
│   ├── service/                  # Business logic layer
│   │   └── user_service.go
│   ├── repository/               # Data access layer (query ke DB)
│   │   └── user_repository.go
│   ├── model/                    # Struct model / entity (mapping tabel DB)
│   │   └── user.go
│   ├── dto/                      # Data Transfer Object (request/response struct)
│   │   └── user_dto.go
│   ├── middleware/                # Custom Echo middleware
│   │   └── auth.go
│   └── cache/                    # Abstraksi Redis cache
│       └── redis.go
├── pkg/
│   ├── database/                 # Setup koneksi Bun + PostgreSQL
│   │   └── postgres.go
│   ├── redis/                    # Setup koneksi Redis client
│   │   └── redis.go
│   └── response/                 # Helper standar HTTP response
│       └── response.go
├── routes/
│   └── routes.go                 # Registrasi semua route Echo
├── test/
│   └── integration/              # Integration test (opsional, terpisah dari unit test)
├── .env.example                  # Template environment variable
├── go.mod
├── go.sum
└── README.md
```

### Penjelasan Setiap Layer

| Layer | Lokasi | Tanggung Jawab |
|---|---|---|
| **Handler** | `internal/handler/` | Menerima HTTP request, validasi input, memanggil service, mengembalikan response. Tidak boleh berisi business logic. |
| **Service** | `internal/service/` | Berisi seluruh business logic. Memanggil satu atau lebih repository. Tidak tahu soal HTTP. |
| **Repository** | `internal/repository/` | Berinteraksi langsung dengan database melalui Bun ORM. Satu repository per entity/tabel. |
| **Model** | `internal/model/` | Definisi struct yang merepresentasikan tabel database. Menggunakan Bun struct tag. |
| **DTO** | `internal/dto/` | Struct untuk request body dan response body. Terpisah dari model agar tidak expose struktur DB. |
| **Cache** | `internal/cache/` | Abstraksi operasi cache (get, set, delete) menggunakan Redis. |
| **Config** | `config/` | Parsing environment variable / config file ke struct Go. |
| **Routes** | `routes/` | Registrasi semua endpoint ke Echo instance. Menghubungkan path → handler. |
| **Pkg** | `pkg/` | Kode yang bersifat *reusable* dan tidak mengandung business logic (koneksi DB, Redis client, response helper). |

---

## 2. Arsitektur Aplikasi

### Alur Request

```
HTTP Request
    │
    ▼
┌──────────┐     ┌──────────┐     ┌──────────────┐     ┌────────────┐
│  Router  │────▶│ Handler  │────▶│   Service    │────▶│ Repository │
│ (Echo)   │     │          │     │              │     │   (Bun)    │
└──────────┘     └──────────┘     └──────────────┘     └────────────┘
                      │                  │                     │
                      │                  │                     ▼
                      │                  │              ┌────────────┐
                      │                  ├─────────────▶│   Cache    │
                      │                  │              │  (Redis)   │
                      ▼                  │              └────────────┘
                 HTTP Response           │
                                         ▼
                                   ┌────────────┐
                                   │ PostgreSQL │
                                   └────────────┘
```

### Prinsip Separation of Concerns

- **Handler** hanya bertanggung jawab untuk: parsing request, validasi input (format/tipe), memanggil service, dan mengembalikan HTTP response.
- **Service** hanya bertanggung jawab untuk: menjalankan business logic, orchestration antar repository/cache, dan mengembalikan hasil atau error.
- **Repository** hanya bertanggung jawab untuk: eksekusi query ke database. Tidak boleh tahu soal HTTP maupun business rule.
- **Cache** bertanggung jawab untuk: menyimpan dan mengambil data dari Redis. Dipanggil oleh service layer.

> **Aturan emas:** Setiap layer hanya boleh berkomunikasi dengan layer di bawahnya melalui **interface**. Handler → Service Interface → Repository Interface.

---

## 3. Standar Penggunaan

### 3.1 Bun ORM

- Setiap model di `internal/model/` menggunakan `bun.BaseModel` sebagai embedded struct.
- Gunakan Bun struct tags (`bun:"table:users"`, `bun:",pk,autoincrement"`, dll.) untuk mapping ke tabel.
- Query ditulis hanya di layer **repository**. Gunakan Bun query builder (`db.NewSelect()`, `db.NewInsert()`, dst.).
- Untuk migrasi, gunakan mekanisme migration bawaan Bun atau tool terpisah — **jangan** buat/alter tabel langsung dari kode aplikasi.

### 3.2 Integrasi PostgreSQL

- Koneksi database di-setup di `pkg/database/postgres.go` dan diinisialisasi di `main.go`.
- Gunakan `sql.Open()` + `bun.NewDB()` untuk membuat instance Bun DB.
- Connection pool dikonfigurasi melalui `sql.DB` (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`).
- DSN / connection string diambil dari config (environment variable), bukan hardcoded.

### 3.3 Integrasi Redis

- Redis client di-setup di `pkg/redis/redis.go` menggunakan library `github.com/redis/go-redis/v9`.
- Abstraksi cache di `internal/cache/` menggunakan interface agar bisa di-mock saat testing.
- Koneksi Redis (host, port, password, db) diambil dari config.

### 3.4 Konvensi Penamaan

| Aspek | Konvensi | Contoh |
|---|---|---|
| File | `snake_case` | `user_handler.go`, `order_service.go` |
| Struct | `PascalCase` | `UserHandler`, `OrderService` |
| Interface | `PascalCase` + deskriptif | `UserRepository`, `OrderService` |
| Method | `PascalCase` (exported) | `FindByID()`, `Create()` |
| Variable | `camelCase` | `userRepo`, `orderSvc` |
| Package | `lowercase`, singular | `handler`, `service`, `repository` |
| Tabel DB | `snake_case`, plural | `users`, `order_items` |
| Environment Variable | `UPPER_SNAKE_CASE` | `DB_HOST`, `REDIS_PORT` |

---

## 4. Testing & Mocking

### 4.1 Desain untuk Testability

Kunci utama agar kode mudah di-test adalah **semua dependency diterima melalui interface, bukan concrete type**.

```
// Contoh konsep (bukan implementasi final):

type UserRepository interface {
    FindByID(ctx context.Context, id int64) (*model.User, error)
    Create(ctx context.Context, user *model.User) error
}

type UserService struct {
    repo UserRepository   // ← interface, bukan *UserRepositoryImpl
}
```

- **Handler** menerima service interface.
- **Service** menerima repository interface dan cache interface.
- **Repository** menerima `*bun.DB` (atau interface wrapper jika diperlukan).

### 4.2 Pendekatan Unit Testing

- **Unit test** ditulis di file `_test.go` berdampingan dengan file yang ditest (contoh: `user_service.go` → `user_service_test.go`).
- Gunakan **mock** untuk semua dependency external (DB, Redis, service lain).
- Library yang direkomendasikan untuk mocking: `github.com/stretchr/testify/mock` atau code generation tool seperti `mockery`.
- Test fokus pada **behavior**, bukan implementasi internal.
- Setiap public method di service dan handler harus memiliki minimal test untuk: happy path, error case, dan edge case.

### 4.3 Struktur Test

```
internal/
├── service/
│   ├── user_service.go
│   └── user_service_test.go      # Unit test + mock
├── handler/
│   ├── user_handler.go
│   └── user_handler_test.go      # Unit test + mock service
├── repository/
│   ├── user_repository.go
│   └── user_repository_test.go   # Test dengan DB (atau mock bun.IDB)
```

- Integration test (yang butuh DB/Redis asli) diletakkan di `test/integration/` dan dijalankan terpisah.

---

## 5. Best Practices

### 5.1 Dependency Injection

- Semua dependency di-inject melalui **constructor function** (`NewUserService(repo UserRepository) *UserService`).
- Wiring dependency dilakukan di `main.go` atau file bootstrap terpisah.
- **Tidak** menggunakan global variable untuk menyimpan instance DB, Redis, atau service.
- Urutan inisialisasi di `main.go`:
  1. Load config
  2. Setup DB connection
  3. Setup Redis connection
  4. Inisialisasi repository
  5. Inisialisasi service (inject repository)
  6. Inisialisasi handler (inject service)
  7. Register routes
  8. Start server

### 5.2 Error Handling

- Gunakan **custom error type** atau sentinel error untuk membedakan jenis error (not found, validation, internal).
- Service layer mengembalikan error yang bermakna secara bisnis.
- Handler layer menerjemahkan error dari service menjadi HTTP status code yang sesuai.
- **Jangan** expose error internal (stack trace, query SQL) ke response API. Log internal error, kembalikan pesan yang aman ke client.
- Gunakan `fmt.Errorf("context: %w", err)` untuk wrapping error agar chain error tetap terjaga.

### 5.3 Config Management

- Semua konfigurasi dibaca dari **environment variable** (12-Factor App).
- Gunakan library seperti `github.com/caarlos0/env` atau `github.com/spf13/viper` untuk parsing env ke struct.
- Sediakan file `.env.example` sebagai dokumentasi variable yang dibutuhkan.
- Config struct di-load sekali di awal dan di-pass ke komponen yang membutuhkan (bukan dibaca ulang di setiap layer).
- Pisahkan config per concern: `DatabaseConfig`, `RedisConfig`, `ServerConfig`.

### 5.4 Middleware

- Gunakan middleware bawaan Echo untuk kebutuhan umum: `Logger`, `Recover`, `CORS`.
- Custom middleware (misalnya auth, rate limiter) diletakkan di `internal/middleware/`.
- Middleware bersifat cross-cutting dan tidak boleh mengandung business logic.

### 5.5 API Response

- Buat format response yang konsisten untuk seluruh endpoint:
  ```json
  {
    "success": true,
    "message": "...",
    "data": { ... }
  }
  ```
- Response error juga menggunakan format yang sama agar client mudah handle.
- Helper response diletakkan di `pkg/response/`.

---

## Ringkasan Teknologi & Library

| Kebutuhan | Library / Tool |
|---|---|
| HTTP Framework | `github.com/labstack/echo/v4` |
| ORM | `github.com/uptrace/bun` |
| Database Driver | `github.com/lib/pq` atau `github.com/jackc/pgx/v5` |
| Redis Client | `github.com/redis/go-redis/v9` |
| Config | `github.com/caarlos0/env` atau `github.com/spf13/viper` |
| Testing | `testing` (stdlib) + `github.com/stretchr/testify` |
| Mocking | `github.com/vektra/mockery` atau `github.com/stretchr/testify/mock` |
| Live Reload (dev) | `github.com/air-verse/air` |

---

> **Langkah selanjutnya:** Gunakan dokumen ini sebagai acuan untuk mulai implementasi. Mulai dari setup koneksi database & redis (`pkg/`), lalu buat model, repository, service, dan handler untuk entity pertama.
