package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gtuk/discordwebhook"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	updateInterval = 60 * time.Second
	dataFilePath   = "./last_earthquakes.json"
)

type earthquake struct {
	Date           string `json:"date"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	Depth          string `json:"depth"`
	Violence       string `json:"violence"`
	Location       string `json:"location"`
	SolutionNature string `json:"solution_nature"`
}

func getLastEarthquakes() ([]earthquake, error) {
	res, err := http.Get("http://www.koeri.boun.edu.tr/scripts/lst8.asp")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var earthquakes []earthquake

	doc.Find("pre").Each(func(i int, s *goquery.Selection) {
		lines := strings.Split(s.Text(), "\n")
		for _, line := range lines[6:] {
			data := strings.Fields(line)
			if len(data) < 11 {
				continue
			}

			last := data[len(data)-1]

			e := earthquake{
				Date:      data[0] + " " + data[1],
				Latitude:  data[2],
				Longitude: data[3],
				Depth:     data[4],
				Violence:  data[6],
				Location:  data[8] + data[9],
			}
			if strings.Contains(last, ")") || strings.Contains(last, "(") {
				e.SolutionNature = data[len(data)-3] + " " + data[len(data)-2] + " " + last
			} else {
				e.SolutionNature = "İlksel"
			}

			earthquakes = append(earthquakes, e)
		}
	})

	return earthquakes, nil
}

func sendWebhook(earthquake *earthquake) error {

	var username = "Deprem Duyurucu"
	var content = "**__Yeni deprem tespit edildi!__**\n\nTarih: **" + earthquake.Date + "**\nKordinat: **" + earthquake.Latitude + "," + earthquake.Longitude + "**\nDerinlik: **" + earthquake.Depth + "**\nKonum: **" +
		earthquake.Location + "**\nDeprem boyutu : **" + earthquake.Violence + "**\nÇözüm Niteliği: **" + earthquake.SolutionNature + "**\n-------------------------------------------"
	var url = os.Getenv("WEBHOOK")

	message := discordwebhook.Message{
		Username: &username,
		Content:  &content,
	}

	err := discordwebhook.SendMessage(url, message)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

var lastEarthquakes []earthquake

func main() {
	for {
		earthquakes, err := getLastEarthquakes()
		if err != nil {
			fmt.Println("Deprem bilgileri alınamadı:", err)
			continue
		}

		for _, earthquake := range earthquakes {
			found := false
			for _, last := range lastEarthquakes {
				if earthquake.Date == last.Date && earthquake.Location == last.Location {
					found = true
					break
				}
			}

			if !found {
				err := sendWebhook(&earthquake)
				if err != nil {
					fmt.Println("Webhook gönderilemedi:", err)
				} else {
					fmt.Println("Webhook başarıyla gönderildi:", earthquake)
					time.Sleep(updateInterval)
				}
			}
		}

		lastEarthquakes = earthquakes
	}

}
