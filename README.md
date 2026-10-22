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
docker compose up -d          # MySQL
cd backend && go run .        # API di :8080
cd frontend && npm install && npm run dev
```

Akun awal: `admin` / `admin123`. Segera ganti di menu Akun (buat akun baru, hapus admin awal).

## Environment variable

Backend (isi di server/docker, bukan di file):

| Nama | Isi | Default |
|---|---|---|
| DB_DSN | `user:password@tcp(host:3306)/rejeki` | root:root@tcp(127.0.0.1:3306)/rejeki |
| JWT_SECRET | teks acak panjang, WAJIB diganti di server | ganti-secret-ini |
| PORT | port API | 8080 |
| GIN_MODE | isi `release` di server | debug |

Frontend: salin `frontend/.env.example` menjadi `frontend/.env`, isi `VITE_API_URL` dengan alamat API
(contoh `https://domain-kamu.com/api`), lalu `npm run build`. Hasilnya di `frontend/dist`.

## Menambah fitur baru

1. Backend: buat controller di `controllers/`, daftarkan route di `routes/`, taruh logika yang dipakai ulang di `services/`.
2. Kalau menu baru: tambah kunci di `services/access.go` (SemuaMenu) dan `config/menu.js` + `pages/pengaturan/Akun.jsx` (KATALOG).
3. Frontend: buat halaman di `pages/`, tambahkan ke `config/menu.js`.
