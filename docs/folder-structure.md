# Struktur folder

Struktur ini dipakai sebagai fondasi untuk semua project API Go baru.

```text
gokit-starter/
├── cmd/
│   └── server/
│       └── main.go
├── docs/
│   ├── brief.md
│   └── folder-structure.md
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── platform/
│   │   ├── database/
│   │   │   └── database.go
│   │   └── logger/
│   │       └── logger.go
│   └── modules/
│       └── <domain>/
│           ├── app/
│           ├── domain/
│           └── infra/
└── ...
```

## Aturan pakai

- `cmd/` hanya untuk entrypoint aplikasi;
- `internal/config` untuk konfigurasi runtime;
- `internal/platform` untuk dependency umum seperti database dan logger;
- `internal/modules` untuk logic per domain;
- `docs/` untuk catatan arsitektur dan keputusan awal.

## Saat project berkembang

```text
internal/
├── app/
├── delivery/
│   ├── handler/
│   ├── middleware/
│   └── response/
├── modules/
│   ├── auth/
│   ├── user/
│   ├── product/
│   └── ...
├── platform/
│   ├── auth/
│   ├── cache/
│   ├── database/
│   ├── logger/
│   ├── queue/
│   └── validation/
└── shared/
    ├── constants/
    ├── errors/
    ├── events/
    └── ids/
```

## Catatan

Jika proyek masih kecil, tetap mulai dari struktur minimal. Tambahkan folder baru hanya ketika memang ada kebutuhan nyata.
