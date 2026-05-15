# Gokit Starter

Gokit Starter adalah boilerplate untuk memulai project API Go dengan struktur yang konsisten, ringan, dan mudah dikembangkan.

## Tujuan repo

- menjadi basis awal untuk project API Go berikutnya;
- menjaga pemisahan yang jelas antara konfigurasi, platform, delivery, dan domain;
- memudahkan cloning dan penyesuaian untuk kebutuhan tiap project;
- menyediakan fondasi standar untuk database, logging, validasi, auth, cache, dan background job;
- mengurangi waktu setup awal sebelum masuk ke fitur bisnis.

## Prinsip utama

- `cmd/` hanya berisi entrypoint;
- `internal/config` menyimpan semua konfigurasi runtime;
- `internal/platform` berisi dependency umum yang dipakai lintas fitur;
- `internal/modules` berisi logic per domain atau feature;
- `internal/delivery` tetap tipis dan fokus ke HTTP;
- konfigurasi dibaca dari environment, bukan hardcode.

## Ruang lingkup awal

- HTTP server;
- config loader;
- logger terstruktur;
- koneksi database;
- schema dan migrasi;
- fondasi untuk auth, cache, queue, dan observability.

## Bukan tujuan utama

- menjadi aplikasi produk yang penuh fitur dari awal;
- menaruh business logic di handler;
- mengikat repo ke satu domain bisnis tertentu;
- membuat struktur yang terlalu kompleks sebelum diperlukan.

## Cara pakai untuk project baru

1. Duplikasi repo ini.
2. Ganti nama module, app, dan default env sesuai project.
3. Tambahkan module yang memang dibutuhkan.
4. Simpan pola arsitektur yang sama supaya semua project tetap konsisten.
