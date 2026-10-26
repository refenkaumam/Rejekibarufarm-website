package controllers

import (
	"math"
	"testing"
	"time"
)

func TestCsvSafe(t *testing.T) {
	cases := map[string]string{
		"=HYPERLINK(\"x\")": "'=HYPERLINK(\"x\")",
		"+1":                "'+1",
		"-2":                "'-2",
		"@SUM(A1)":          "'@SUM(A1)",
		"Setoran kas":       "Setoran kas",
		"":                  "",
	}
	for in, want := range cases {
		if got := csvSafe(in); got != want {
			t.Errorf("csvSafe(%q) = %q, mau %q", in, got, want)
		}
	}
}

func TestValidTgl(t *testing.T) {
	if !validTgl(time.Now().Format("2006-01-02")) {
		t.Error("tanggal hari ini ditolak")
	}
	for _, s := range []string{"1999-12-31", "2026-13-01", "abc", "2026-1-5", time.Now().AddDate(2, 0, 0).Format("2006-01-02")} {
		if validTgl(s) {
			t.Errorf("tanggal %q seharusnya ditolak", s)
		}
	}
}

func TestValidBulan(t *testing.T) {
	if !validBulan("2026-10") || validBulan("2026-1") || validBulan("x' OR 1=1") {
		t.Error("validasi bulan salah")
	}
}

func TestInRange(t *testing.T) {
	if inRange(math.NaN(), 0, 1) || inRange(math.Inf(1), 0, 1e12) || inRange(-1, 0, 1) || !inRange(0.5, 0, 1) {
		t.Error("validasi rentang angka salah")
	}
}
