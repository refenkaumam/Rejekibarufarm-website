-- Migrasi aditif untuk fitur keamanan (tidak menghapus/mengubah data lama).
-- Backend menjalankan perubahan yang sama secara otomatis saat start (database.Migrate),
-- jadi file ini hanya perlu dijalankan manual bila ingin menyiapkan database lebih dulu.
-- Untuk database lama: jalankan hanya bagian yang belum ada (kolom token_version bisa sudah ditambah backend).

-- Versi sesi: dinaikkan saat logout/ganti password agar token lama tidak berlaku
ALTER TABLE users ADD COLUMN token_version INT NOT NULL DEFAULT 0;

-- Jejak audit: siapa, kapan, aksi, data mana, nilai sebelum & sesudah, IP (tanpa password/token)
CREATE TABLE IF NOT EXISTS audit_log (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  user_id BIGINT NULL,
  username VARCHAR(50) NULL,
  ip VARCHAR(45) NULL,
  aksi VARCHAR(30) NOT NULL,
  tabel VARCHAR(30) NULL,
  row_id BIGINT NULL,
  sebelum MEDIUMTEXT NULL,
  sesudah MEDIUMTEXT NULL,
  KEY idx_audit_at (at),
  KEY idx_audit_row (tabel, row_id)
);
