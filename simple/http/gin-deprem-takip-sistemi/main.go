package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

type road struct {
	ID     int
	Name   string
	Reason string
}

type earthquake struct {
	ID             int
	Date           string `json:"date"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	Depth          string `json:"depth"`
	Violence       string `json:"violence"`
	Location       string `json:"location"`
	SolutionNature string `json:"solution_nature"`
}

func fetchRoads() []road {
	res, _ := http.Get("https://www.kgm.gov.tr/Sayfalar/KGM/SiteTr/YolDanisma/TrafigeKapaliYollar.aspx")
	defer res.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(res.Body)

	var roads []road

	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}
		id, _ := strconv.Atoi(strings.TrimSpace(s.Find("td:nth-child(1)").Text()))
		name := strings.TrimSpace(s.Find("td:nth-child(6)").Text())
		reason := strings.TrimSpace(s.Find("td:nth-child(7)").Text())
		roads = append(roads, road{ID: id, Name: name, Reason: reason})
	})

	return roads
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("src/*.tmpl")
	r.Static("/static", "./static")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Anasayfa",
		})
	})

	r.GET("/konumum", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "konum.tmpl", gin.H{
			"title": "Konumunuz",
		})
	})

	r.GET("/harita", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "harita.tmpl", gin.H{
			"title": "Deprem haritası",
		})
	})

	r.GET("/deprem/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		i, err := strconv.Atoi(id)
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID must be an integer"})
			return
		}

		res, err := http.Get("http://www.koeri.boun.edu.tr/scripts/lst8.asp")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": err.Error()})
			return
		}
		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": err.Error()})
			return
		}

		var earthquakes []earthquake
		var idCount int = 0
		doc.Find("pre").Each(func(i int, s *goquery.Selection) {
			lines := strings.Split(s.Text(), "\n")
			for _, line := range lines[6:] {
				data := strings.Fields(line)
				if len(data) < 11 {
					continue
				}

				last := data[len(data)-1]

				e := earthquake{
					ID:        idCount,
					Date:      data[0] + " " + data[1],
					Latitude:  data[2],
					Longitude: data[3],
					Depth:     data[4],
					Violence:  data[6],
					Location:  data[8] + data[9],
				}
				idCount++
				if strings.Contains(last, ")") || strings.Contains(last, "(") {
					e.SolutionNature = data[len(data)-3] + " " + data[len(data)-2] + " " + last
				} else {
					e.SolutionNature = "İlksel"
				}

				earthquakes = append(earthquakes, e)
			}
		})

		if i >= len(earthquakes) {
			ctx.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Not found"})
			return
		}

		ctx.HTML(http.StatusOK, "deprem.tmpl", gin.H{
			"earthquake": earthquakes[i],
			"ID":         earthquakes[i].ID,
			"Date":       earthquakes[i].Date,
			"Latitude":   earthquakes[i].Latitude,
			"Longitude":  earthquakes[i].Longitude,
			"Depth":      earthquakes[i].Depth,
			"Violence":   earthquakes[i].Violence,
			"Location":   earthquakes[i].Location,
		})
	})

	r.GET("/yakinimi-bul", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "yakinimi-bul.tmpl", gin.H{
			"title": "Yakınımı bul",
		})
	})

	r.GET("/kapali-yollar", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "kapali-yollar.tmpl", gin.H{
			"title": "Kapalı yollar",
		})
	})

	r.GET("/api/earthquake/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		i, err := strconv.Atoi(id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID must be an integer"})
			return
		}

		res, err := http.Get("http://www.koeri.boun.edu.tr/scripts/lst8.asp")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var earthquakes []earthquake
		var idCount int = 0
		doc.Find("pre").Each(func(i int, s *goquery.Selection) {
			lines := strings.Split(s.Text(), "\n")
			for _, line := range lines[6:] {
				data := strings.Fields(line)
				if len(data) < 11 {
					continue
				}

				last := data[len(data)-1]

				e := earthquake{
					ID:        idCount,
					Date:      data[0] + " " + data[1],
					Latitude:  data[2],
					Longitude: data[3],
					Depth:     data[4],
					Violence:  data[6],
					Location:  data[8] + data[9],
				}
				idCount++
				if strings.Contains(last, ")") || strings.Contains(last, "(") {
					e.SolutionNature = data[len(data)-3] + " " + data[len(data)-2] + " " + last
				} else {
					e.SolutionNature = "İlksel"
				}

				earthquakes = append(earthquakes, e)
			}
		})

		if i >= len(earthquakes) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}

		ctx.JSON(http.StatusOK, earthquakes[i])
	})

	r.GET("/api/last/earthquakes", func(ctx *gin.Context) {
		res, err := http.Get("http://www.koeri.boun.edu.tr/scripts/lst8.asp")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var earthquakes []earthquake
		var id int = 0
		doc.Find("pre").Each(func(i int, s *goquery.Selection) {
			lines := strings.Split(s.Text(), "\n")
			for _, line := range lines[6:] {
				data := strings.Fields(line)
				if len(data) < 11 {
					continue
				}

				last := data[len(data)-1]

				e := earthquake{
					ID:        id,
					Date:      data[0] + " " + data[1],
					Latitude:  data[2],
					Longitude: data[3],
					Depth:     data[4],
					Violence:  data[6],
					Location:  data[8] + data[9],
				}
				id++
				if strings.Contains(last, ")") || strings.Contains(last, "(") {
					e.SolutionNature = data[len(data)-3] + " " + data[len(data)-2] + " " + last
				} else {
					e.SolutionNature = "İlksel"
				}

				earthquakes = append(earthquakes, e)
			}
		})
		ctx.JSON(http.StatusOK, earthquakes)
	})

	r.GET("/api/roads", func(c *gin.Context) {
		roads := fetchRoads()
		c.JSON(200, roads)
	})

	r.Run(":5000")
}
