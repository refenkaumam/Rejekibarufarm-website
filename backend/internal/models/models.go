package models

type Sapi struct {
	ID         int64  `json:"id"`
	Nomor      string `json:"nomor"`
	Nama       string `json:"nama"`
	TglLahir   string `json:"tgl_lahir"`
	JumlahAnak int    `json:"jumlah_anak"`
	Status     string `json:"status"`
	UmurBulan  int    `json:"umur_bulan"`
}

type Pakan struct {
	ID         int64   `json:"id"`
	Tgl        string  `json:"tgl"`
	Jenis      string  `json:"jenis"`
	Arah       string  `json:"arah"`
	Jumlah     float64 `json:"jumlah"`
	Satuan     string  `json:"satuan"`
	HargaTotal float64 `json:"harga_total"`
	Keterangan string  `json:"keterangan"`
}

type Baris struct {
	Label  string  `json:"label"`
	Jumlah float64 `json:"jumlah"`
}

type LR struct {
	Pendapatan      []Baris `json:"pendapatan"`
	Biaya           []Baris `json:"biaya"`
	TotalPendapatan float64 `json:"total_pendapatan"`
	TotalBiaya      float64 `json:"total_biaya"`
	Laba            float64 `json:"laba"`
}
