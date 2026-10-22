-- Peternakan om: skema awal (MySQL 8)
CREATE TABLE users (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(50) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  -- super_admin, owner, peternakan, melk, back_office
  role VARCHAR(20) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 1. Rekap sapi (umur dihitung otomatis dari tgl_lahir)
CREATE TABLE sapi (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  nomor VARCHAR(30) UNIQUE NOT NULL,
  nama VARCHAR(100) NOT NULL,
  tgl_lahir DATE NOT NULL,
  jumlah_anak INT NOT NULL DEFAULT 0,
  status ENUM('laktasi','kering','dara','pedet') NOT NULL,
  investor_id BIGINT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Stok & kebutuhan pakan (masuk = pembelian -> otomatis kas keluar)
CREATE TABLE pakan_mutasi (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  tgl DATE NOT NULL,
  jenis ENUM('rumput','konsentrat','tebon_jagung','mineral') NOT NULL,
  arah ENUM('masuk','keluar') NOT NULL,
  jumlah DECIMAL(12,2) NOT NULL,
  satuan VARCHAR(10) NOT NULL DEFAULT 'kg',
  harga_total DECIMAL(15,2) NOT NULL DEFAULT 0,
  keterangan VARCHAR(255)
);

-- 3. Produksi susu harian
CREATE TABLE produksi_susu (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  tgl DATE NOT NULL,
  sesi ENUM('pagi','sore') NOT NULL,
  liter DECIMAL(10,2) NOT NULL,
  UNIQUE KEY uq_tgl_sesi (tgl, sesi)
);

-- 4. Penjualan susu (koperasi / melk_rbf), uang masuk = liter x harga
CREATE TABLE penjualan_susu (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  tgl DATE NOT NULL,
  tujuan ENUM('koperasi','melk_rbf') NOT NULL,
  liter DECIMAL(10,2) NOT NULL,
  harga_per_liter DECIMAL(12,2) NOT NULL
);

-- 5. Investor
CREATE TABLE investor (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  nama VARCHAR(100) NOT NULL,
  kontak VARCHAR(100),
  jumlah_invest DECIMAL(15,2) NOT NULL,
  mulai DATE NOT NULL,
  selesai DATE NOT NULL
);
ALTER TABLE sapi ADD FOREIGN KEY (investor_id) REFERENCES investor(id);

-- Melk by RBF: keju & susu masuk/keluar
CREATE TABLE melk_keju (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  tgl DATE NOT NULL,
  arah ENUM('masuk','keluar') NOT NULL,
  jumlah_gram DECIMAL(12,2) NOT NULL,
  harga DECIMAL(15,2) NOT NULL,
  pembeli VARCHAR(100) NULL  -- diisi jika keluar (dijual ke siapa)
);
CREATE TABLE melk_susu (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  tgl DATE NOT NULL,
  arah ENUM('masuk','keluar') NOT NULL,
  liter DECIMAL(10,2) NOT NULL,
  harga DECIMAL(15,2) NOT NULL,
  pembeli VARCHAR(100) NULL
);

-- Kas: satu tabel untuk 3 unit (kandang, melk, back_office)
CREATE TABLE kas (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  tgl DATE NOT NULL,
  unit ENUM('kandang','melk','back_office') NOT NULL,
  arah ENUM('masuk','keluar') NOT NULL,
  -- kandang masuk: kas, pembelian_susu, topup_bo, investor
  -- kandang keluar: gaji, beli_pakan, operasional, bulanan_investor, lainnya
  kategori VARCHAR(30) NOT NULL,
  jumlah DECIMAL(15,2) NOT NULL,
  keterangan VARCHAR(255),
  ref_tabel VARCHAR(30) NULL,  -- sumber otomatis (pakan_mutasi, penjualan_susu, dst)
  ref_id BIGINT NULL
);

CREATE TABLE kas_awal_bo (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  tgl DATE NOT NULL,
  jumlah DECIMAL(15,2) NOT NULL
);
