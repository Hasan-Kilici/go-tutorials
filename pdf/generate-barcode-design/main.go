package main

import (
	"fmt"
	"image/png"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
	"github.com/signintech/gopdf"
)

func main() {

	r := gin.Default()
	r.LoadHTMLGlob("src/*.tmpl")
	r.Static("/static", "./static")
	r.Static("/pdf", "./pdf")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "newBarcode.tmpl", gin.H{
			"title": "Barkod oluştur",
		})
	})

	r.POST("/newBarcode/", func(ctx *gin.Context) {
		//Barkod
		barcodeXInput := ctx.PostForm("barkodX")
		barcodeX, err := strconv.ParseFloat(barcodeXInput, 64)
		if err != nil {
			fmt.Println("X Konumu floata çevrilemedi")
		}
		barcodeYInput := ctx.PostForm("barkodY")
		barcodeY, err := strconv.ParseFloat(barcodeYInput, 64)
		if err != nil {
			fmt.Println("Y Konumu floata çevrilemedi")
		}
		barcodeWidthInput := ctx.PostForm("barkodGenislik")
		barcodeWidth, err := strconv.ParseInt(barcodeWidthInput, 10, 64)
		if err != nil {
			fmt.Println("Barkod genişliği int64'e çevrilemedi")
		}
		barcodeHeightInput := ctx.PostForm("barkodUzunluk")
		barcodeHeight, err := strconv.ParseInt(barcodeHeightInput, 10, 64)
		if err != nil {
			fmt.Println("Barkod uzunluğu int64'e çevrilemedi")
		}
		//Ürün Fiyatı
		priceXInput := ctx.PostForm("fiyatX")
		priceX , err := strconv.ParseFloat(priceXInput, 64)
		if err != nil {
			fmt.Println("Ürün adının X konumu float64'e çevrilemedi")
		}
		priceYInput := ctx.PostForm("fiyatY")
		priceY , err := strconv.ParseFloat(priceYInput, 64)
		if err != nil {
			fmt.Println("Ürün adının Y konumu float64'e çevrilemedi")	
		}
		priceFontSizeInput := ctx.PostForm("fiyatFontBoyutu")
		priceFontSize, err := strconv.ParseFloat(priceFontSizeInput, 64)
		if err != nil {
			fmt.Println("Ürün adının Font boyutu float64'e çevrilemedi")	
		}
		//Ürün adı
		productXInput := ctx.PostForm("urunX")
		productX , err := strconv.ParseFloat(productXInput, 64)
		if err != nil {
			fmt.Println("Ürün adının X konumu float64'e çevrilemedi")
		}
		productYInput := ctx.PostForm("urunY")
		productY , err := strconv.ParseFloat(productYInput, 64)
		if err != nil {
			fmt.Println("Ürün adının Y konumu float64'e çevrilemedi")	
		}
		productFontSizeInput := ctx.PostForm("urunFontBoyutu")
		productFontSize, err := strconv.ParseFloat(productFontSizeInput, 64)
		if err != nil {
			fmt.Println("Ürün adının Font boyutu float64'e çevrilemedi")	
		}
		//Barkod boyutu (8 cm genişlik - 4 cm uzunluk)
		pdf := gopdf.GoPdf{}
		pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
		pdf.AddPageWithOption(gopdf.PageOption{
			PageSize: &gopdf.Rect{
				W: 8.0 * 28.3464567,
				H: 4.0 * 28.3464567,
			},
		})

		// sayfa içeriğini ekle
		pdf.AddTTFFont("Arial", "./tff/arial.ttf")
		pdf.SetFont("Times", "", 14)

		enc := oned.NewEAN13Writer()
		img, err := enc.Encode("400638133393", gozxing.BarcodeFormat_EAN_13, int(barcodeWidth), int(barcodeHeight), nil)
		if err != nil {
			fmt.Println("Barkoda çevrilemedi")
		}
		file, err := os.Create("./data/barcode.png")
		if err != nil {
			fmt.Println("Barcode.png oluşturulamadı")
		}
		defer file.Close()
		//Barkod
		_ = png.Encode(file, img)
		pdf.Image("./data/barcode.png", barcodeX, float64(barcodeY), nil)
		//Barkod'un metni
		pdf.SetXY(float64(barcodeX + float64(barcodeX / 4)), float64(barcodeY+float64(barcodeHeight)))
		pdf.SetFont("Arial", "", 10)
		pdf.Text("400638133393")
		//Ürün fiyatı
		pdf.SetFont("Arial", "", priceFontSize)
		pdf.SetXY(priceX, priceY)
		pdf.Text("31 TL")
		//Ürün adı
		pdf.SetFont("Arial", "", productFontSize)
		pdf.SetXY(productX, productY)
		pdf.Text("Ürün adı")

		pdf.WritePdf("./pdf/example.pdf")
		ctx.Redirect(http.StatusFound, "/")
	})
	r.Run(":5000")
}