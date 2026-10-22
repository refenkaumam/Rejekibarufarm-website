package services

import (
	"database/sql"
	"strings"

	"rejeki-baru-farm/internal/database"
)

// Kunci menu (sama dengan kunci di frontend)
var SemuaMenu = []string{"sapi", "pakan", "produksi", "penjualan", "investor", "kas",
	"msusu", "mkeju", "mrekap", "mkas", "bo", "bokas"}

func DefaultMenus(role string) []string {
	switch role {
	case "super_admin", "owner":
		return SemuaMenu
	case "peternakan":
		return SemuaMenu[:6]
	case "melk":
		return SemuaMenu[6:10]
	case "back_office":
		return SemuaMenu[10:]
	}
	return []string{}
}

// NULL = ikut default role, "" = tidak ada menu, selain itu = daftar menu pilihan
func Efektif(role string, m sql.NullString) []string {
	if role == "super_admin" || !m.Valid {
		return DefaultMenus(role)
	}
	if m.String == "" {
		return []string{}
	}
	return strings.Split(m.String, ",")
}

func MenusOf(username string) (string, []string, bool) {
	var role string
	var m sql.NullString
	if database.DB.QueryRow("SELECT role,menus FROM users WHERE username=?", username).Scan(&role, &m) != nil {
		return "", nil, false
	}
	return role, Efektif(role, m), true
}

func CleanMenus(in []string) string {
	ok := map[string]bool{}
	for _, k := range SemuaMenu {
		ok[k] = true
	}
	out := []string{}
	for _, k := range in {
		if ok[k] {
			out = append(out, k)
		}
	}
	return strings.Join(out, ",")
}
