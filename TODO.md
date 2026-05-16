# 🚀 Gokit Starter DX Improvements TODO

Daftar perbaikan Developer Experience (DX) untuk menyempurnakan boilerplate.

- [x] **1. Versioned Database Migrations (Ent + Atlas)**
    - [x] Setup Atlas CLI integration dengan Ent.
    - [x] Pisahkan logic `Schema.Create` (auto) menjadi `Migration` (versioned).
    - [x] Tambahkan `task db:diff` di Taskfile untuk generate file SQL migrasi.
    - [x] Tambahkan `task db:migrate` untuk apply migrasi ke database.
    - [x] Update dokumentasi cara kelola database.

- [ ] **2. Fail-Fast Configuration (Viper + Validator)**
    - [ ] Tambahkan struct tags `validate:"required"` pada struct `Config`.
    - [ ] Integrasikan `go-playground/validator` di dalam `config.LoadConfig()`.
    - [ ] Pastikan aplikasi langsung error jika `.env` tidak lengkap saat startup.

- [ ] **3. Dependency Graph Visualization (Uber Fx)**
    - [ ] Tambahkan hook `fx.Visualize()` untuk generate DOT graph.
    - [ ] Tambahkan perintah di Taskfile untuk ekspor visualisasi dependency.
    - [ ] Integrasikan dengan flag `APP_DEBUG` atau perintah CLI khusus.

- [ ] **4. Module Generator Scaffolding**
    - [ ] Buat template folder untuk module baru (`domain`, `app`, `infra`).
    - [ ] Tambahkan `task new:module` yang menerima parameter `name`.
    - [ ] Otomasi pembuatan boilerplate code untuk module baru.

---
*Gunakan `task test` dan `task lint` setiap kali menyelesaikan satu poin di atas.*
