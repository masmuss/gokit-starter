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
│   ├── app/
│   │   ├── fx.go
│   │   └── router.go
│   ├── config/
│   │   └── config.go
│   ├── delivery/
│   │   ├── handler/
│   │   ├── middleware/
│   │   └── response/
│   ├── platform/
│   │   ├── auth/
│   │   ├── cache/
│   │   ├── database/
│   │   ├── logger/
│   │   └── validation/
│   ├── shared/
│   │   ├── errors/
│   │   └── event/
│   └── modules/
│       └── <domain>/
│           ├── app/
│           ├── domain/
│           └── infra/
└── ...
```

## Aturan pakai

- `cmd/` hanya untuk entrypoint aplikasi;
- `internal/app` tempat bootstrapping DI (Uber Fx) dan routing;
- `internal/config` untuk konfigurasi runtime;
- `internal/platform` untuk dependency umum seperti database, cache, dan logger;
- `internal/shared` untuk komponen yang mendukung komunikasi antar module (Event Bus) atau utilitas global (Shared Errors);
- `internal/modules` untuk logic per domain;
- `docs/` untuk catatan arsitektur dan keputusan awal.

## Saat project berkembang

```text
internal/
├── app/
├── delivery/
│   ├── handler/
│   │   ├── auth_handler.go
│   │   └── health_handler.go
│   ├── middleware/
│   └── response/
│       └── response.go
├── modules/
│   ├── auth/
│   │   ├── app/
│   │   ├── domain/
│   │   └── infra/
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
