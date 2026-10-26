# Laporan Perbaikan Keamanan — Rejeki Baru Farm

- **Tanggal:** 2026-10-26
- **Cakupan:** backend Go (Gin + `database/sql` + `go-sql-driver/mysql`, tanpa ORM), frontend React 19 + Vite (SPA tanpa router, halaman dipilih lewat state), `docker-compose.yml`.
- **Ringkasan:** 28 temuan. 23 diperbaiki penuh, 2 diperbaiki tapi masih perlu tindakan manual di server, 3 diterima sebagai risiko karena menyangkut alur bisnis.
- **Verifikasi:** `go build ./...`, `go vet ./...`, `go test ./...` (test baru), `govulncheck`, `npm run lint`, `npm run build`, `npm audit` lolos. Selain itu ada uji end-to-end ke MySQL **sementara** yang terpisah (port 3307, data di folder temp) dan uji browser nyata dengan CSP produksi. Database Laragon milikmu (port 3306) tidak disentuh.

---

## 0. Peta sistem (ringkas)

| Hal | Temuan |
|---|---|
| Endpoint | 46 route di `/api/*`. Semua selain `POST /api/login` lewat `Auth()`. Data per menu lewat `Akses(menu)`, kelola akun lewat `Allow("super_admin")`, ekspor lewat `Allow(super_admin, owner)` + `Akses(bo)`. Baru: `POST /api/logout`, `PUT /api/me/password`, `PUT /api/users/:id/password`. |
| Login & sesi (sebelum) | JWT HS256 24 jam disimpan di `localStorage`, dikirim lewat header `Authorization`. Role diambil dari isi token. |
| Login & sesi (sesudah) | JWT 12 jam di **cookie `rbf_session` HttpOnly + SameSite** (Secure di produksi). Identitas, role, dan menu dibaca ulang dari database setiap request. Token CSRF (HMAC dari ID sesi) wajib dikirim di header `X-CSRF-Token` untuk POST/PUT/DELETE. |
| Role | `super_admin`, `owner`, `peternakan`, `melk`, `back_office`. Hak akses per menu bisa diatur super admin. Satu peternakan (single-tenant, tidak ada data milik user lain). |
| Penyajian produksi | Repo tidak punya konfigurasi produksi. README hanya menyarankan `npm run build`. Sekarang ada `deploy/nginx.conf.example` yang menyajikan `frontend/dist` sebagai file statis dan mem-proxy `/api` (bukan `vite`/`vite preview`). |
| Origin | Dev: beda origin tapi satu site (`localhost:5173` → `localhost:8080`). Produksi yang disarankan: satu origin (`/api` di domain yang sama). |
| CORS | Sebelum: `*`. Sesudah: hanya origin di `ALLOWED_ORIGINS`, dengan credentials. |
| Database | `go-sql-driver/mysql`, semua query parameterized. Sekarang ada timeout koneksi dan batas pool. |
| Docker | Hanya MySQL. Sebelum: `root/root`, port `3306` terbuka ke semua interface. |
| `.env` | `backend/.env` ada tapi **kosong (0 byte)**, dan sebelumnya tidak di-ignore. Isi `frontend/.env*` hanya `VITE_API_URL` (bukan rahasia). |

---

## 1. Daftar temuan

Status: **Diperbaiki** / **Perlu tindakan manual** / **Diterima sebagai risiko**.

### Kritis

| # | Temuan | Lokasi | Penjelasan | Status |
|---|---|---|---|---|
| 1 | JWT secret jatuh ke nilai bawaan yang ada di repo | `backend/internal/config/config.go`, `backend/main.go` | Kalau `JWT_SECRET` kosong, server memakai `ganti-secret-ini`, sehingga siapa pun bisa membuat token super_admin. Sekarang di produksi server **menolak start** bila secret < 32 karakter atau bernilai bawaan. Di development dipakai secret acak sementara. | Diperbaiki (+ rotasi secret di server, lihat §4) |

### Tinggi

| # | Temuan | Lokasi | Penjelasan | Status |
|---|---|---|---|---|
| 2 | Akun bawaan `admin/admin123` & tidak ada fitur ganti password | `services/auth.go` (`SeedAdmin`) | Akun baru sekarang memakai `ADMIN_INITIAL_PASSWORD` atau password acak (ditampilkan sekali di log). Ditambah endpoint ganti password sendiri dan reset oleh super admin. Server mencatat peringatan bila `admin` masih login dengan `admin123`. **Akun yang sudah terlanjur ada di database tidak berubah otomatis.** | Perlu tindakan manual (ganti password, §4) |
| 3 | Login tanpa pembatasan percobaan | `routes/routes.go`, `services/ratelimit.go` | Terkunci 15 menit setelah 5 kali gagal per IP+username, atau 30 kali gagal per IP. Ditambah batas 600 request/menit/IP untuk seluruh API dan `limit_req` di contoh Nginx. | Diperbaiki |
| 4 | MySQL `root/root` dan port 3306 terbuka | `docker-compose.yml` | Port sekarang hanya `127.0.0.1`, password dibaca dari `.env`, ada user khusus aplikasi. Volume Docker lama tetap memakai password lama. | Diperbaiki + Perlu tindakan manual |
| 5 | Token login di `localStorage` (mudah dicuri lewat XSS) | `frontend/src/api/*.js`, `RekapBO.jsx` | Token dipindah ke cookie HttpOnly yang diset backend. Frontend memakai `credentials: 'include'` + header CSRF. Diverifikasi di browser: `document.cookie` kosong, tidak ada token di `localStorage`. | Diperbaiki |
| 6 | Tidak ada jejak audit, catatan keuangan bisa dihapus/diubah tanpa jejak | semua controller | Tabel baru `audit_log` mencatat login berhasil/gagal/diblokir, logout, ganti/reset password, tambah/ubah/hapus semua data, perubahan akun & akses menu, dan ekspor. Isinya: user, waktu, IP, aksi, ID data, nilai sebelum & sesudah (tanpa password/token). | Diperbaiki |
| 7 | Sesi tidak bisa dicabut; role dipercaya dari isi token | `middleware/auth.go` | Role & menu dibaca dari database setiap request. Ada `token_version` per user: logout, ganti password, atau reset password langsung membatalkan token lama. User yang dihapus langsung kehilangan akses. | Diperbaiki |
| 8 | `backend/.env` tidak di-ignore git | root | `.gitignore` baru di root (+ `frontend/.gitignore` diperbarui), `.env.example` tanpa nilai asli. | Diperbaiki + cek riwayat git (§4) |

### Sedang

| # | Temuan | Lokasi | Penjelasan | Status |
|---|---|---|---|---|
| 9 | Formula injection pada ekspor CSV | `controllers/backoffice.go` | Teks yang diawali `= + - @` (tab/CR) diberi awalan `'`. Nama file di header kini dari tanggal tervalidasi dan diberi tanda kutip. | Diperbaiki |
| 10 | Error database diabaikan → respons "ok" padahal gagal, transaksi bisa tersimpan sebagian | semua controller | Semua `Exec`/`Commit` dicek lewat helper `withTx`. Kalau satu langkah gagal, semua di-rollback. Data yang tidak ada → 404. | Diperbaiki |
| 11 | Validasi input backend lemah | semua controller, `controllers/helpers.go` | Format & rentang tanggal, batas angka (uang ≤ 1e12, liter ≤ 1e6, kg ≤ 1e7, gram ≤ 1e9), jumlah anak 0–30, tanggal lahir 1990–hari ini, panjang teks sesuai kolom, enum, `bulan`/`arah`/`dari`/`sampai` di query, ID path harus angka positif, investor harus ada, periode investor masuk akal. | Diperbaiki |
| 12 | CORS `*` | `middleware/cors.go` | Hanya origin dari `ALLOWED_ORIGINS`. Preflight dari origin lain → 403. Request pengubah data dari origin asing → 403. | Diperbaiki |
| 13 | Dependensi rentan: stdlib go1.26.0 (26 celah terpanggil) & `quic-go` v0.59.0 | `backend/go.mod` | `toolchain go1.26.10` + `quic-go v0.59.1`. `govulncheck`: 0 celah terpanggil. Sisa 1: `x/crypto/openpgp` (GO-2026-5932, tidak dipakai kode, belum ada perbaikan). `npm audit`: 0. | Diperbaiki (openpgp: diterima) |
| 14 | Batas antar-unit: staf kandang bisa mencatat `topup_bo` yang mengurangi saldo Back Office, dan menghapus penjualan ke Melk ikut menghapus data Melk | `kas_kandang.go`, `penjualan.go` | Ini alur bisnis yang sudah ada. Mengubahnya berarti mengubah cara kerja user. Sekarang semua aksi itu tercatat di audit log. | Diterima sebagai risiko |
| 15 | Hapus/ubah data boleh dilakukan siapa pun yang punya menu terkait | routes | Membatasi hapus hanya untuk owner akan mengubah apa yang biasa dilakukan staf. Mitigasinya audit log (nilai sebelum tersimpan, jadi bisa dipulihkan manual). | Diterima sebagai risiko |

### Rendah

| # | Temuan | Lokasi | Penjelasan | Status |
|---|---|---|---|---|
| 16 | Pesan error SQL internal dikirim ke client | semua controller | Detail error hanya ditulis ke log server. Client menerima "Terjadi kesalahan server". | Diperbaiki |
| 17 | Gin mode debug, semua proxy dipercaya, tanpa timeout, tanpa batas body, panic tanpa JSON | `routes.go`, `main.go`, `middleware/security.go` | Mode release di produksi, `TRUSTED_PROXIES` eksplisit, timeout server (header 5s, baca 15s, tulis 60s, idle 60s), body maks 1 MB, recovery yang mengembalikan JSON generik, graceful shutdown. | Diperbaiki |
| 18 | Tidak ada header keamanan | `middleware/security.go`, `deploy/nginx.conf.example` | API: `nosniff`, `X-Frame-Options DENY`, CSP `default-src 'none'`, `Referrer-Policy`, `Permissions-Policy`, `Cache-Control: no-store`, HSTS (produksi+HTTPS). Frontend: contoh Nginx dengan CSP ketat, sudah diuji di browser tanpa pelanggaran. | Diperbaiki (Nginx: perlu dipasang manual) |
| 19 | Enumerasi username lewat beda waktu respons login | `services/auth.go` | User yang tidak ada tetap menjalankan bcrypt dummy. Pesan error tetap generik. | Diperbaiki |
| 20 | bcrypt cost 10, error hash diabaikan | `services/auth.go`, `controllers/user.go` | Cost 12. Hash lama otomatis dinaikkan saat user berhasil login (terverifikasi: `$2a$10$` → `$2a$12$`). Password 10–72 karakter. | Diperbaiki |
| 21 | Login CSRF lewat form/text biasa | `middleware/security.go` | Body POST/PUT/DELETE wajib `application/json` (selain itu 415). | Diperbaiki |
| 22 | Mass assignment (`SaveSapi` bind ke struct model) | `controllers/sapi.go` | Diganti struct input khusus. Semua handler lain sudah memakai struct input sendiri. | Diperbaiki |
| 23 | Super admin terakhir bisa terhapus | `controllers/user.go` | Ditolak dengan pesan jelas. Pengecekan hapus diri sendiri kini memakai ID dari sesi. | Diperbaiki |
| 24 | `Migrate()` menelan semua error | `database/database.go` | Cek kolom lewat `information_schema`. Server berhenti dengan pesan jelas bila skema belum diimpor. | Diperbaiki |
| 25 | Password akun awal ditulis ke log | `services/auth.go` | Hanya bila dibuat acak, ditampilkan sekali. Bila dari env, tidak ditampilkan. | Diperbaiki |
| 26 | Source map / font inline melanggar CSP | `frontend/vite.config.js` | `sourcemap: false` eksplisit. Font tidak di-inline sebagai `data:`, sehingga lolos `font-src 'self'`. | Diperbaiki |
| 27 | Stok pakan bisa negatif | `controllers/pakan.go` | Pencatatan pemakaian sebelum pembelian mungkin memang dipakai di lapangan, jadi tidak diblokir. | Diterima sebagai risiko |
| 28 | Tidak ada backup database | — | Tidak ada skrip/jadwal backup di repo. | Perlu tindakan manual |

**Sudah dicek dan aman:** SQL injection (semua parameterized; satu-satunya konkatenasi memakai konstanta), XSS (tidak ada `dangerouslySetInnerHTML`/`innerHTML`; React meng-escape semua data), open redirect, upload file (fitur tidak ada), path traversal (tidak ada akses file), IDOR (single-tenant, akses dibatasi per menu), `password_hash` tidak pernah dikirim ke client maupun ke audit log (dicek: 0 baris).

---

## 2. File yang diubah / dibuat (untuk review)

**Backend — diubah**
- `backend/main.go` — server dengan timeout, `HOST`, graceful shutdown, migrasi/seed dengan error jelas
- `backend/go.mod`, `backend/go.sum` — `toolchain go1.26.10`, `quic-go v0.59.1`, `go mod tidy`
- `backend/internal/config/config.go` — konfigurasi env + validasi produksi + pembaca `.env`
- `backend/internal/database/database.go` — timeout & pool, migrasi aditif, tabel `audit_log`
- `backend/internal/middleware/auth.go` — sesi dari cookie + DB, CSRF, role/menu dari DB
- `backend/internal/middleware/cors.go` — allowlist origin
- `backend/internal/routes/routes.go`, `routes/users.go` — middleware keamanan, endpoint baru
- `backend/internal/services/auth.go` — JWT (iss/jti/ver), cookie, CSRF, login aman, ganti/reset password, seed
- `backend/internal/services/access.go` — hapus `MenusOf` (tidak dipakai lagi), dedupe menu
- `backend/internal/controllers/` — `auth.go`, `user.go`, `sapi.go`, `pakan.go`, `produksi.go`, `penjualan.go`, `investor.go`, `kas_kandang.go`, `melk_kas.go`, `melk_susu.go`, `melk_keju.go`, `melk_rekap.go`, `backoffice.go` (validasi, transaksi, audit, error generik). **Query bisnis, perhitungan, dan bentuk respons tidak berubah.**

**Backend — baru**
- `backend/internal/middleware/security.go` — header keamanan, batas body, JSON wajib, cek origin, rate limit, recovery
- `backend/internal/services/ratelimit.go`, `services/audit.go`
- `backend/internal/controllers/helpers.go` — validasi & transaksi bersama
- Test: `services/auth_test.go`, `controllers/helpers_test.go`, `routes/routes_test.go`
- `backend/migrations/002_keamanan.sql` — dokumentasi migrasi aditif (backend menjalankannya otomatis)
- `backend/.env.example`, `backend/Dockerfile` (multi-stage, distroless, non-root), `backend/.dockerignore`

**Frontend — diubah**
- `frontend/src/api/client.js` — `credentials: 'include'`, header CSRF, tanpa header `Authorization`
- `frontend/src/api/auth.js` — simpan token CSRF (sessionStorage), logout memanggil backend, hapus sisa `token` lama
- `frontend/src/pages/backoffice/RekapBO.jsx` — unduh CSV pakai cookie (satu baris)
- `frontend/vite.config.js` — `sourcemap: false`, font tidak di-inline
- `frontend/.gitignore` — abaikan `.env*`

**Root**
- Diubah: `docker-compose.yml`, `README.md` (bagian menjalankan & env)
- Baru: `.gitignore`, `.env.example`, `deploy/nginx.conf.example`, `SECURITY_REPORT.md`

**Tidak diubah:** `internal/models/models.go`, `routes/backoffice.go`, `routes/melk.go`, `routes/peternakan.go`, `services/laporan.go`, `migrations/001_init_peternakan.sql`, tampilan/komponen frontend, dan isi `backend/.env`.

> Catatan: working directory juga masih berisi perubahan redesign tampilan dan `SECURITY_AUDIT.md` dari pekerjaan sebelumnya yang belum di-commit. Pisahkan saat commit bila perlu.

---

## 3. Perubahan yang berdampak pada perilaku

| Perubahan | Dampak | Cara deploy |
|---|---|---|
| Token pindah ke cookie | **Semua user login ulang sekali** setelah deploy. Token lama di `localStorage` tidak berlaku lagi. | Deploy backend + frontend bersamaan. |
| Lama sesi 12 jam (sebelumnya 24 jam) | Login ulang lebih sering. | Atur `SESSION_TTL` (15m–168h) bila perlu. |
| Logout = keluar dari semua perangkat milik user itu | Juga terjadi saat ganti/reset password. | — |
| CORS & cek origin | Frontend di origin yang tidak terdaftar akan ditolak (403). | Isi `ALLOWED_ORIGINS` dengan origin frontend, **termasuk bila satu domain**. |
| Penguncian login | 5 kali gagal (per IP+username) atau 30 kali per IP → terkunci 15 menit. Batas global 600 request/menit/IP. Penghitung di memori, reset saat restart. | Di balik proxy, isi `TRUSTED_PROXIES` supaya IP asli terbaca. |
| Password baru minimal 10 karakter (sebelumnya 6) | Hanya saat membuat akun atau ganti password. Password lama tetap bisa dipakai login. | — |
| Validasi input lebih ketat | Nilai tidak wajar ditolak (400): tanggal transaksi sebelum 2000 atau lebih dari 1 tahun ke depan, tanggal lahir sapi di masa depan, jumlah anak > 30, angka di luar batas. | — |
| Data yang tidak ada → 404 | Dulu hapus/ubah ID yang tidak ada tetap "ok". | — |
| Body harus JSON (415), body > 1 MB ditolak | Frontend sudah selalu mengirim JSON. | — |
| Tidak bisa menghapus super admin terakhir | — | — |
| Backend berhenti saat start bila konfigurasi produksi kurang, atau skema 001 belum diimpor | Dulu jalan diam-diam lalu error saat dipakai. | Set env produksi (§4). |
| `DB_DSN` bawaan development menjadi `root@tcp(127.0.0.1:3306)/rejeki` (tanpa password, cocok untuk Laragon) | Dulu `root:root@...`. Pengguna Docker lokal perlu mengisi `DB_DSN`. | `backend/.env`. |
| Migrasi otomatis aditif | Saat start: kolom `users.token_version` (default 0) dan tabel `audit_log` dibuat bila belum ada. Data lama tidak diubah. Diuji pada salinan skema lama. | Tidak perlu langkah manual. |
| Ekspor CSV | Teks berawalan `= + - @` diawali `'`. Ekspor tercatat di audit log. | — |
| `docker-compose.yml` butuh file `.env` | `docker compose up` gagal dengan pesan jelas bila password belum diisi. | Salin `.env.example` → `.env`. |
| Toolchain Go 1.26.10 | `go build` pertama kali otomatis mengunduh toolchain (butuh internet), atau pasang Go ≥ 1.26.10. | — |
| Endpoint baru | `POST /api/logout`, `PUT /api/me/password` (`{password_lama, password_baru}`), `PUT /api/users/:id/password` (`{password}`, super admin). **Belum ada tampilan untuk ganti password** (sesuai aturan tidak mengubah tampilan). | Lihat §4 cara memakainya. |

---

## 4. Tindakan manual untukmu

### a. Secret yang wajib dirotasi (di server, jangan di repo)
1. **`JWT_SECRET`**: buat baru minimal 32 karakter acak. Kalau server pernah berjalan dengan nilai bawaan `ganti-secret-ini`, anggap semua token yang pernah terbit bisa dipalsukan.
   - Linux: `openssl rand -base64 48`
   - PowerShell: `[Convert]::ToBase64String((1..48 | % { Get-Random -Max 256 }))`
2. **Password MySQL `root`**: kalau pernah dipakai `root/root` (bawaan compose lama), ganti, lalu buat user khusus aplikasi dengan hak `SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX` di database `rejeki` saja, dan pakai user itu di `DB_DSN`.
   - Untuk volume Docker yang sudah ada, env baru tidak mengubah password lama. Jalankan sendiri: `ALTER USER 'root'@'%' IDENTIFIED BY '<baru>';`
3. **Password akun `admin`**: kalau masih `admin123`, ganti segera (lihat butir d).

### b. Environment variable produksi (backend)
```
APP_ENV=production
HOST=127.0.0.1
PORT=8080
DB_DSN=<user_app>:<password>@tcp(127.0.0.1:3306)/rejeki
JWT_SECRET=<acak ≥ 32 karakter>
ALLOWED_ORIGINS=https://<domain-frontend>
TRUSTED_PROXIES=127.0.0.1
COOKIE_SECURE=true
COOKIE_SAMESITE=lax
```
Frontend dibuild dengan `VITE_API_URL=/api` (satu domain). Kalau frontend dan API ada di domain berbeda, isi `VITE_API_URL` dengan URL API dan `COOKIE_SAMESITE=none` (wajib HTTPS).

### c. Server, reverse proxy, firewall
- Pasang reverse proxy dengan HTTPS sesuai `deploy/nginx.conf.example`. Header CSP di sana sudah diuji dengan hasil build.
- Sajikan `frontend/dist` sebagai file statis. **Jangan** menjalankan `vite`/`vite preview` di produksi.
- Firewall: tutup port 3306 dan 8080 dari luar. Hanya 80/443 yang terbuka.
- Jalankan backend sebagai user non-root (atau pakai `backend/Dockerfile`).
- Siapkan **backup database** terjadwal (mis. `mysqldump` harian), dienkripsi, dan disimpan di luar folder web.

### d. Mengganti password admin (belum ada tampilan)
Login di aplikasi, buka DevTools → Console, lalu jalankan. Sesuaikan URL: `/api` kalau satu domain, `http://localhost:8080/api` kalau dev.
```js
fetch('/api/me/password', { method: 'PUT', credentials: 'include',
  headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': sessionStorage.getItem('rbf-csrf') },
  body: JSON.stringify({ password_lama: 'PASSWORD_LAMA', password_baru: 'PASSWORD_BARU_MIN_10' }) }).then(r => r.json()).then(console.log)
```
Kalau mau, aku bisa menambahkan form ganti password di halaman Akun (ini perubahan tampilan, jadi tidak kukerjakan tanpa izin).

### e. Cek riwayat git secara manual (tidak kulakukan)
Aku tidak menjalankan perintah git apa pun di tugas ini. Silakan periksa sendiri apakah secret pernah ter-commit:
```
git log --all --full-history -- "*.env" "**/.env"
git log --all -p | findstr /i "password secret jwt_secret dsn"
```
Atau pakai `gitleaks detect`. Commit lama berisi nilai bawaan `root:root`, `admin123`, dan `ganti-secret-ini`. Ketiganya sudah dianggap bocor dan masuk daftar rotasi di atas. Kalau ditemukan secret asli lain di riwayat, rotasi nilainya; menghapus dari riwayat saja tidak cukup.

### f. Lain-lain
- `backend/.env` saat ini kosong. Isi berdasarkan `backend/.env.example` (file ini sekarang diabaikan git).
- Audit log terus bertambah. Tentukan kebijakan retensi, misalnya arsip tahunan.

---

## 5. Konfirmasi

- **Alur bisnis tidak berubah:** query perhitungan (laba rugi, saldo, stok, rekap), relasi otomatis antar-tabel (pakan → kas, penjualan → kas/Melk, investor → kas), dan bentuk respons JSON yang dipakai frontend tetap sama. Respons login/`me` hanya **ditambah** field `csrf`, dan field `token` tidak lagi dikirim di body. Tampilan tidak berubah (diverifikasi lewat screenshot). Perubahan perilaku yang disengaja tercantum di §3.
- **Uji yang dijalankan:**
  - `go build ./...`, `go vet ./...`, `go test ./...` (3 paket) lolos.
  - `govulncheck`: 0 celah terpanggil.
  - `npm run lint`, `npm run build` lolos (0 source map); `npm audit`: 0 kerentanan.
  - Uji end-to-end API dengan curl ke MySQL sementara terpisah: migrasi aditif, upgrade hash, cookie, CSRF, origin, validasi, CSV, ganti password, logout, penguncian login, audit log.
  - Uji browser nyata (login → tambah & hapus sapi → logout) di bawah CSP produksi: tanpa error konsol.
- **Git:** tidak ada perintah git yang dijalankan selama tugas ini, dan folder `.git` tidak disentuh. Sebagai catatan jujur: di sesi sebelumnya (audit pertama & redesign) aku sempat menjalankan perintah git **baca-saja** (`git ls-files`, `git log`, `git status`) sebelum aturan ini diberikan.
- **Database:** tidak ada perintah destruktif ke databasemu. Semua uji memakai instance MySQL sementara dari binary Laragon dengan data directory di folder temp (port 3307), lalu dimatikan. MySQL Laragon di port 3306 tidak disentuh.
