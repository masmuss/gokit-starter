# Code Review: Gokit Starter Boilerplate

Secara keseluruhan ini adalah boilerplate yang solid dengan arsitektur yang jelas. Berikut review menyeluruhnya.

---

## ✅ Yang Sudah Baik

- Clean Architecture + Modular Monolith diterapkan konsisten
- Dependency injection via Uber Fx terstruktur rapi
- Domain layer murni, tidak ada ketergantungan ke infrastruktur
- Test coverage cukup baik (unit + integration + handler test)
- Penggunaan `slog` sebagai logger standar library sudah tepat
- Versioned migrations via Atlas adalah pilihan yang matang

---

## 🔴 Critical Issues

### 1. Race condition pada Event Bus

`internal/shared/event/event.go`

```go
// MASALAH: errCh dibuat dengan kapasitas len(handlers),
// tapi goroutine bisa jalan setelah wg.Wait() selesai
// dan close(errCh) dipanggil → write to closed channel → panic

wg.Wait()
close(errCh)

for err := range errCh {  // ini aman, tapi ada edge case
```

Masalah sebenarnya: jika semua goroutine selesai sebelum `close()`, tidak ada masalah. Tapi logika pengambilan error hanya mengambil **satu** error pertama dan mengabaikan sisanya — ini silent failure.

```go
// PERBAIKAN:
func (b *InternalBus) Publish(ctx context.Context, e Event) error {
    b.mu.RLock()
    handlers, ok := b.subscribers[e.Name]
    b.mu.RUnlock()

    if !ok {
        return nil
    }

    var (
        wg   sync.WaitGroup
        mu   sync.Mutex
        errs []error
    )

    for _, h := range handlers {
        wg.Add(1)
        go func(handler Handler) {
            defer wg.Done()
            if err := handler(ctx, e); err != nil {
                mu.Lock()
                errs = append(errs, err)
                mu.Unlock()
            }
        }(h)
    }

    wg.Wait()
    return errors.Join(errs...)
}
```

---

### 2. CI workflow menggunakan versi actions yang tidak valid

`.github/workflows/ci.yml`

```yaml
# SALAH: versi ini tidak ada
- uses: actions/checkout@v6 # latest stable adalah v4
- uses: actions/setup-go@v6 # latest stable adalah v5
- uses: golangci/golangci-lint-action@v9 # latest adalah v6
```

Ini akan membuat CI gagal total saat dijalankan.

---

### 3. Database password exposed di DSN log

`internal/platform/database/database.go`

```go
// DSN dengan password akan muncul di error message jika koneksi gagal
dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
    cfg.Database.Username,
    cfg.Database.Password,  // ← password plaintext di string yang bisa ter-log
    ...
)

if err = stdDB.Ping(); err != nil {
    return nil, fmt.Errorf("failed to ping database: %w", err)
    // error bisa menyertakan DSN tergantung driver
}
```

```go
// PERBAIKAN: Gunakan pgxpool atau sembunyikan credential di error
dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
    cfg.Database.Username,
    cfg.Database.Password,
    cfg.Database.Host,
    cfg.Database.Port,
    cfg.Database.Database,
)

if err = stdDB.Ping(); err != nil {
    // Jangan expose DSN, buat error message yang aman
    return nil, fmt.Errorf("failed to ping database at %s:%d: %w",
        cfg.Database.Host, cfg.Database.Port, err)
}
```

---

## 🟡 Medium Issues

### 4. `BindJSON` bisa panic pada request body nil

`internal/platform/validation/validation.go`

```go
func BindJSON(r *http.Request, dst any) error {
    decoder := json.NewDecoder(r.Body)  // panic jika r.Body == nil
    decoder.DisallowUnknownFields()
    // ...
    // Decode kedua hanya untuk cek trailing data — ini fragile
    if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
        return errors.Join(ErrInvalidJSON, errors.New("unexpected trailing data"))
    }
```

Masalah lain: `DisallowUnknownFields()` bisa menyebabkan error yang membingungkan client jika ada field tambahan. Pertimbangkan apakah ini perilaku yang diinginkan.

```go
func BindJSON(r *http.Request, dst any) error {
    if r.Body == nil {
        return errors.Join(ErrInvalidJSON, errors.New("empty request body"))
    }
    defer r.Body.Close()

    // Batasi ukuran body untuk mencegah DoS
    limited := http.MaxBytesReader(nil, r.Body, 1<<20) // 1MB
    decoder := json.NewDecoder(limited)

    if err := decoder.Decode(dst); err != nil {
        return errors.Join(ErrInvalidJSON, err)
    }
    return nil
}
```

---

### 5. Redis connection error tidak di-handle dengan benar di Fx

`internal/app/fx.go` + `internal/platform/cache/cache.go`

```go
// cache.go: NewRedisClient mengembalikan error
func NewRedisClient(cfg *config.Config) (*redis.Client, error) { ... }

// fx.go: Ini sudah benar, Fx akan handle error-nya
// TAPI: jika Redis down, seluruh aplikasi tidak bisa start
// Untuk cache yang seharusnya opsional, ini terlalu strict
```

Jika Redis adalah optional dependency (cache), pertimbangkan graceful degradation:

```go
// Buat Redis optional dengan fallback no-op cache
func NewRedisClientOptional(cfg *config.Config, log *slog.Logger) *redis.Client {
    client, err := tryConnectRedis(cfg)
    if err != nil {
        log.Warn("redis unavailable, cache disabled", "error", err)
        return nil
    }
    return client
}
```

---

### 6. Tidak ada request body size limit di router

`internal/app/router.go`

```go
// Tidak ada middleware untuk membatasi ukuran request body
// Rentan terhadap memory exhaustion attack
r.Use(middleware.RequestID)
r.Use(middleware.RealIP)
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)
```

```go
// TAMBAHKAN:
r.Use(middleware.RequestID)
r.Use(middleware.RealIP)
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)
r.Use(middleware.CleanPath)
r.Use(middleware.Timeout(30 * time.Second))  // request timeout
// Body size limit handled per-handler atau global
```

---

### 7. `organizationCode` bisa menghasilkan collision

`internal/modules/auth/infra/repository.go`

```go
func organizationCode(name string) string {
    base := normalizeCode(name)
    if base == "" {
        base = "org"
    }

    suffix := strings.ReplaceAll(uuid.NewString(), "-", "")[:6]
    // Hanya 6 karakter hex = 16^6 = ~16 juta kemungkinan
    // Dengan volume tinggi, collision bisa terjadi
    // Dan tidak ada retry logic jika constraint error terjadi
```

Tidak ada unique constraint pada `code` di schema, tapi logika ini mengasumsikan uniqueness. Tambahkan unique constraint di Ent schema atau tambahkan retry logic.

---

### 8. `slog.HandlerOptions` dengan `nil` writer di test

`internal/delivery/handler/auth_handler_test.go`

```go
// Ini akan panic: nil writer tidak valid untuk TextHandler
slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{}))
```

```go
// PERBAIKAN:
slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
```

---

## 🟢 Minor / Suggestions

### 9. Inkonsistensi naming konvensi di `compose.yml`

```yaml
# compose.yml menggunakan postgres:18-alpine
# ci.yml menggunakan postgres:15-alpine
# Seharusnya konsisten untuk menghindari behavior yang berbeda
```

---

### 10. `RegisterRequest` field `organization_name` validate tag salah

`internal/delivery/handler/auth_handler.go`

```go
type RegisterRequest struct {
    Name             string `json:"name"              validate:"required"`
    Email            string `json:"email"             validate:"required,email"`
    Password         string `json:"password"          validate:"required,min=8"`
    OrganizationName string `json:"organization_name" validate:"omitempty"`
    // validate:"omitempty" tidak berguna sendirian, field ini opsional secara default
    // Jika ingin validasi saat diisi: validate:"omitempty,min=3"
}
```

---

### 11. Tidak ada pagination pada query potensial

`internal/modules/auth/infra/repository.go`

Untuk `FindByEmail` dan `FindByID` ini oke. Tapi boilerplate tidak menyediakan helper pagination yang akan dibutuhkan semua modul baru. Pertimbangkan tambahkan di `internal/shared/`:

```go
// internal/shared/pagination/pagination.go
type Params struct {
    Page    int
    PerPage int
}

func (p Params) Offset() int { return (p.Page - 1) * p.PerPage }
func (p Params) Limit() int  { return p.PerPage }
```

---

### 12. JWT TTL config membingungkan

`internal/config/config.go` + `internal/modules/auth/app/service.go`

```go
// Config: JWTTTL dalam menit
JWTTTL int `mapstructure:"jwt_ttl"` // docs: "60" → 60 menit

// JWTManager: ttl dalam Duration
NewJWTManager(secret, issuer string, ttl time.Duration)
// Di NewJWTManagerFromConfig: time.Duration(cfg.Auth.JWTTTL)*time.Minute ✓

// Tapi di Service.NewFromConfig:
return New(repo, hasher, tokens, cfg.Auth.JWTTTL*60)
// cfg.Auth.JWTTTL = 60 (menit) * 60 = 3600 detik
// expiresIn dalam response = 3600 (detik) — ini BENAR

// Tapi env var namanya AUTH_JWT_TTL dan defaultnya "60"
// Tidak jelas 60 apa. Tambahkan komentar di .env.example
```

Tambahkan di `.env.example`:

```bash
AUTH_JWT_TTL=60  # dalam menit
```

---

### 13. Tidak ada rate limiting

Router tidak memiliki rate limiting sama sekali. Untuk boilerplate production-ready, ini penting:

```go
// Gunakan chi rate limiter atau implementasi sederhana
import "golang.org/x/time/rate"
// atau gunakan middleware third-party seperti go-chi/httprate
```

---

### 14. `scripts/generator/main.go` tidak mengikuti template auth module

Generator menghasilkan `domain.Repository` interface di dalam package `domain`, tapi pattern di auth module menempatkan interface `Repository` di package `app`. Ini inkonsistensi arsitektur yang akan membingungkan developer baru.

---

## Ringkasan Prioritas

| #    | Issue                             | Severity    | Effort        |
| ---- | --------------------------------- | ----------- | ------------- |
| 1    | Race condition event bus          | 🔴 Critical | Rendah        |
| 2    | CI workflow versi invalid         | 🔴 Critical | Sangat Rendah |
| 3    | Password exposed di DSN error     | 🔴 Critical | Rendah        |
| 4    | BindJSON nil body + no size limit | 🟡 Medium   | Rendah        |
| 5    | Redis hard dependency             | 🟡 Medium   | Sedang        |
| 6    | No request body size limit        | 🟡 Medium   | Rendah        |
| 7    | Organization code collision       | 🟡 Medium   | Sedang        |
| 8    | nil writer di test                | 🟡 Medium   | Sangat Rendah |
| 9-14 | Minor improvements                | 🟢 Low      | Bervariasi    |
