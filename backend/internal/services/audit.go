package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// Actor adalah pelaku sebuah aksi (diambil dari sesi di server, bukan dari client).
type Actor struct {
	UserID   int64
	Username string
	IP       string
}

type Execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// Kolom yang disimpan di audit log per tabel. password_hash sengaja tidak pernah ikut.
var snapCols = map[string]string{
	"sapi":           "id,nomor,nama,tgl_lahir,jumlah_anak,status,investor_id",
	"pakan_mutasi":   "id,tgl,jenis,arah,jumlah,satuan,harga_total,keterangan",
	"produksi_susu":  "id,tgl,sesi,liter",
	"penjualan_susu": "id,tgl,tujuan,liter,harga_per_liter",
	"investor":       "id,nama,kontak,jumlah_invest,mulai,selesai",
	"melk_keju":      "id,tgl,arah,jumlah_gram,harga,pembeli",
	"melk_susu":      "id,tgl,arah,liter,harga,pembeli,ref_id",
	"kas":            "id,tgl,unit,arah,kategori,jumlah,keterangan,ref_tabel,ref_id",
	"kas_awal_bo":    "id,tgl,jumlah",
	"users":          "id,username,role,menus,token_version",
}

// SnapshotWhere membaca baris yang cocok untuk dicatat sebagai nilai sebelum/sesudah.
// `where` selalu konstanta dari kode, nilai dari user lewat args (parameterized).
func SnapshotWhere(q Querier, table, where string, args ...any) ([]map[string]any, error) {
	cols, ok := snapCols[table]
	if !ok {
		return nil, fmt.Errorf("tabel audit tidak dikenal: %s", table)
	}
	rows, err := q.Query("SELECT "+cols+" FROM "+table+" WHERE "+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	names, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	out := []map[string]any{}
	for rows.Next() {
		vals := make([]any, len(names))
		ptrs := make([]any, len(names))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		m := map[string]any{}
		for i, n := range names {
			if b, ok := vals[i].([]byte); ok {
				m[n] = string(b)
			} else {
				m[n] = vals[i]
			}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Snapshot membaca satu baris berdasarkan id; nil kalau tidak ada.
func Snapshot(q Querier, table string, id int64) (map[string]any, error) {
	rows, err := SnapshotWhere(q, table, "id=?", id)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0], nil
}

func encode(v any) any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok && m == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return string(b)
}

func nullable(s string, max int) any {
	if s == "" {
		return nil
	}
	if len(s) > max {
		s = s[:max]
	}
	return s
}

// Audit mencatat siapa, kapan, aksi, data mana, dan nilai sebelum/sesudah.
func Audit(ex Execer, a Actor, aksi, tabel string, rowID int64, before, after any) error {
	var uid, rid any
	if a.UserID > 0 {
		uid = a.UserID
	}
	if rowID > 0 {
		rid = rowID
	}
	_, err := ex.Exec(`INSERT INTO audit_log (user_id,username,ip,aksi,tabel,row_id,sebelum,sesudah) VALUES (?,?,?,?,?,?,?,?)`,
		uid, nullable(a.Username, 50), nullable(a.IP, 45), aksi, nullable(tabel, 30), rid, encode(before), encode(after))
	return err
}
