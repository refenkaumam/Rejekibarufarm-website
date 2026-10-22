package routes

import (
	"github.com/gin-gonic/gin"

	"rejeki-baru-farm/internal/controllers"
	"rejeki-baru-farm/internal/middleware"
)

func registerPakanProduksi(api *gin.RouterGroup) {
	g := api.Group("")
	g.GET("/pakan", middleware.Akses("pakan"), controllers.ListPakan)
	g.POST("/pakan", middleware.Akses("pakan"), controllers.AddPakan)
	g.DELETE("/pakan/:id", middleware.Akses("pakan"), controllers.DelPakan)
	g.GET("/pakan/stok", middleware.Akses("pakan"), controllers.StokPakan)
	g.GET("/produksi", middleware.Akses("produksi"), controllers.ListProduksi)
	g.POST("/produksi", middleware.Akses("produksi"), controllers.SaveProduksi)
	g.GET("/produksi/rekap", middleware.Akses("produksi"), controllers.RekapProduksi)
}

func registerPenjualanInvestorKas(api *gin.RouterGroup) {

	g := api.Group("")
	g.GET("/penjualan", middleware.Akses("penjualan"), controllers.ListPenjualan)
	g.POST("/penjualan", middleware.Akses("penjualan"), controllers.AddPenjualan)
	g.DELETE("/penjualan/:id", middleware.Akses("penjualan"), controllers.DelPenjualan)
	g.GET("/investor", middleware.Akses("investor"), controllers.ListInvestor)
	g.POST("/investor", middleware.Akses("investor"), controllers.SaveInvestor)
	g.PUT("/investor/:id", middleware.Akses("investor"), controllers.SaveInvestor)
	g.DELETE("/investor/:id", middleware.Akses("investor"), controllers.DelInvestor)
	g.GET("/investor-bayar", middleware.Akses("investor"), controllers.ListBayar)
	g.POST("/investor-bayar", middleware.Akses("investor"), controllers.AddBayar)
	g.DELETE("/investor-bayar/:id", middleware.Akses("investor"), controllers.DelBayar)
	g.GET("/kas", middleware.Akses("kas"), controllers.ListKas)
	g.POST("/kas", middleware.Akses("kas"), controllers.AddKas)
	g.DELETE("/kas/:id", middleware.Akses("kas"), controllers.DelKas)
}
