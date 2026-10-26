# Rejeki Baru Farm - Website Pembukuan Peternakan

Stack: Go (Gin) + MySQL, React + Vite.

## Struktur folder

```
backend/
  main.go                 titik masuk (baca config, konek DB, jalankan server)
  migrations/             skema database (.sql)
  internal/
    config/               konfigurasi dari environment variable
    database/             koneksi MySQL dan migrasi kolom tambahan
    models/               struct data bersama
    middleware/           CORS, cek login (Auth), cek role (Allow), cek menu (Akses)
    routes/               daftar endpoint per modul
    controllers/          handler HTTP per modul (baca request, panggil DB, kirim response)
    services/             logika bisnis (login/token, hak akses menu, laba rugi, saldo)
frontend/src/
  api/                    client.js (semua request ke backend), auth.js (login, sesi)
  components/             komponen bersama (Sidebar, Field, Stat, Empty)
  config/menu.js          daftar menu dan halaman
  pages/                  satu file per halaman, dikelompokkan: peternakan, melk, backoffice, pengaturan, auth
  utils/format.js         format tanggal, angka, rupiah
  styles/index.css        gaya tampilan
```

## Menjalankan (development)

```
# MySQL: Laragon (root tanpa password, impor backend/migrations/001_init_peternakan.sql ke database `rejeki`)
#   atau Docker: salin .env.example -> .env, isi password, lalu `docker compose up -d`
cd backend && cp .env.example .env && go run .   # API di :8080, tabel tambahan dibuat otomatis
cd frontend && npm install && npm run dev
```

Akun awal (`admin`) hanya dibuat bila tabel `users` kosong. Password-nya dari `ADMIN_INITIAL_PASSWORD`,
atau dibuat acak dan ditampilkan sekali di log backend. Segera ganti password setelah login pertama.

## Environment variable

Backend membaca environment variable dan file `backend/.env` (lihat `backend/.env.example`, jangan di-commit).

| Nama | Isi | Default development |
|---|---|---|
| APP_ENV | `production` di server (konfigurasi wajib divalidasi) | development |
| DB_DSN | `user:password@tcp(host:3306)/rejeki` | root@tcp(127.0.0.1:3306)/rejeki |
| JWT_SECRET | minimal 32 karakter acak, WAJIB di produksi | acak sementara |
| ALLOWED_ORIGINS | origin frontend, pisah koma | http://localhost:5173 |
| TRUSTED_PROXIES | IP reverse proxy | (kosong) |
| COOKIE_SECURE / COOKIE_SAMESITE | `true` / `lax` di produksi HTTPS | false / lax |
| SESSION_TTL | lama sesi login | 12h |
| ADMIN_INITIAL_PASSWORD | password admin pertama (min. 10 karakter) | (acak) |
| HOST / PORT | alamat & port API | semua interface / 8080 |

Frontend: salin `frontend/.env.example` menjadi `frontend/.env`, isi `VITE_API_URL` dengan alamat API
(contoh `/api` bila satu domain), lalu `npm run build`. Hasil `frontend/dist` disajikan web server
(contoh: `deploy/nginx.conf.example`). Jangan memakai `vite` / `vite preview` di produksi.

## Deploy (Docker + Cloudflare Tunnel)

`docker-compose.prod.yml` menjalankan backend dan Nginx (frontend + proxy `/api`) tanpa membuka port ke host.
Database memakai container MariaDB yang sudah ada di server (network `DB_NETWORK`, bawaan `db-net`).
Container `web` bergabung ke network `private-net` milik cloudflared dengan nama `rejeki-web`.

1. Lewat phpMyAdmin server: buat database `rejeki` (collation `utf8mb4_unicode_ci`), impor
   `backend/migrations/001_init_peternakan.sql` **atau** hasil export database lama, lalu buat user `rejeki_app`
   dengan hak `SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX, REFERENCES` hanya pada database `rejeki`.
   Kolom & tabel tambahan (`menus`, `token_version`, `audit_log`) dibuat otomatis oleh backend.
2. Di server, salin `.env.example` menjadi `.env` (folder root) lalu isi bagian Produksi.
3. Jalankan:
   ```
   docker compose -f docker-compose.prod.yml up -d --build
   docker compose -f docker-compose.prod.yml logs -f backend
   ```
4. Cloudflare Zero Trust → Tunnel → Public Hostname: Service **HTTP**, URL **`rejeki-web:8080`**.
5. Di dashboard Cloudflare matikan **Rocket Loader** dan **Email Address Obfuscation** (keduanya menyisipkan script yang diblokir CSP).

Update: tarik kode terbaru lalu jalankan ulang perintah `up -d --build`. Data tersimpan di MariaDB server (backup lewat phpMyAdmin/mysqldump).

## Menambah fitur baru

1. Backend: buat controller di `controllers/`, daftarkan route di `routes/`, taruh logika yang dipakai ulang di `services/`.
2. Kalau menu baru: tambah kunci di `services/access.go` (SemuaMenu) dan `config/menu.js` + `pages/pengaturan/Akun.jsx` (KATALOG).
3. Frontend: buat halaman di `pages/`, tambahkan ke `config/menu.js`.
