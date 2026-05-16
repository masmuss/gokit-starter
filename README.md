# Gokit Starter

Gokit Starter adalah boilerplate modern untuk membangun API menggunakan Go. Proyek ini dirancang dengan prinsip **Clean Architecture** dan **Modular Monolith**, fokus pada pemisahan kepentingan, testability, dan skalabilitas.

## 🚀 Tech Stack

- **Framework**: [Go-Chi](https://github.com/go-chi/chi) (Router)
- **Dependency Injection**: [Uber Fx](https://github.com/uber-go/fx)
- **ORM**: [Ent](https://entgo.io/) (Entity Framework for Go)
- **Caching**: [Redis](https://redis.io/) (via [go-redis](https://github.com/redis/go-redis))
- **Communication**: Internal In-Memory Event Bus
- **Validation**: [Go-Playground Validator](https://github.com/go-playground/validator)
- **Logging**: [Slog](https://pkg.go.dev/log/slog) (Standard Library)
- **Testing**: [Testify](https://github.com/stretchr/testify) & [Mockery](https://github.com/vektra/mockery) (Automated Mocks)
- **Documentation**: [Swagger/OpenAPI](https://github.com/swaggo/swag)
- **Otomasi**: [Taskfile](https://taskfile.dev/) & [Lefthook](https://github.com/evilmartians/lefthook)

## 📁 Struktur Folder

Struktur ini mengikuti pola yang memisahkan infrastruktur, transport layer, dan business logic:

```text
├── cmd/                # Entry points aplikasi
├── docs/               # Dokumentasi arsitektur & Swagger UI
├── internal/
│   ├── app/            # Bootstrapping (Fx Modules, Router setup)
│   ├── config/         # Konfigurasi runtime
│   ├── delivery/       # Transport layer (HTTP Handlers, Middleware)
│   ├── modules/        # Domain business logic (Modular Monolith)
│   │   └── auth/       # Contoh module: Auth
│   │       ├── app/    # Service / Use Cases
│   │       ├── domain/ # Kontrak, Entitas, & Error Domain
│   │       └── infra/  # Implementasi Repository (Database)
│   ├── platform/       # Cross-cutting concerns (DB, Logger, Redis, Auth Utils)
│   └── shared/         # Komponen bersama (Event Bus, Standardized Errors)
├── test/
│   └── mocks/          # Mock yang dihasilkan secara otomatis
└── taskfile.yml        # Task runner (pengganti Makefile)
```

## 🛠️ Persiapan Pengembangan

Pastikan Anda sudah menginstal:
- Go 1.21+
- [Task](https://taskfile.dev/installation/) (`brew install go-task`)
- Docker (untuk database PostgreSQL dan Redis)

### Langkah Awal

1. Clone repository:
   ```bash
   git clone git@github.com:masmuss/gokit-starter.git
   ```
2. Salin environment variables:
   ```bash
   cp .env.example .env
   ```
3. Jalankan infrastruktur (PostgreSQL & Redis):
   ```bash
   docker compose up -d
   ```
4. Jalankan aplikasi dengan hot-reload:
   ```bash
   task server
   ```

## ⌨️ Perintah Task (Taskfile)

Gunakan `task` untuk menjalankan perintah umum:

- `task server`: Menjalankan server dengan `air` (hot-reload).
- `task test`: Menjalankan semua unit test.
- `task mocks`: Menghasilkan mock otomatis menggunakan Mockery.
- `task generate`: Menghasilkan kode Ent dari skema.
- `task db:diff name=...`: Mendeteksi perubahan skema dan men-generate file SQL migrasi baru.
- `task db:migrate`: Menjalankan semua file migrasi SQL yang belum terpakai ke database.
- `task db:clean`: Reset database development (Docker) ke kondisi bersih.
- `task docs:generate`: Menghasilkan dokumentasi Swagger/OpenAPI.
- `task lint`: Menjalankan linter (golangci-lint).
- `task format`: Merapikan format kode Go.

## 🗄️ Manajemen Database

Proyek ini menggunakan **Versioned Migrations** melalui **Atlas**. Jangan melakukan perubahan database secara manual atau mengandalkan auto-migration di produksi.

### Alur Perubahan Skema:
1. Modifikasi skema di `internal/platform/database/ent/schema/`.
2. Jalankan `task generate` untuk memperbarui kode Go.
3. Jalankan `task db:diff name=deskripsi_perubahan` untuk membuat file migrasi `.sql`.
4. Review file SQL yang dihasilkan di folder `internal/platform/database/migrations/`.
5. Jalankan `task db:migrate` untuk menerapkan perubahan ke database lokal Anda.

## 🧪 Testing & Mocks

Proyek ini menggunakan **Mockery** untuk mempermudah unit testing. Jika Anda menambahkan interface baru di layer `app`, jalankan perintah berikut untuk memperbarui mock:

```bash
task mocks
```

Mocks akan tersedia di folder `test/mocks/` dan siap digunakan dalam file `*_test.go`.

## 📖 Kontribusi

1. Buat branch baru dari `dev`.
2. Pastikan `task test` dan `task lint` lulus sebelum membuat Pull Request.
3. Ikuti standar penamaan dan arsitektur yang sudah ada di `internal/modules`.

---

_Gokit Starter - Built for developers who love clean and maintainable code._
