package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	r := gin.Default()
	r.LoadHTMLFiles("index.html")
  r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/ws", func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer ws.Close()
		clients[ws] = true
		for {
			var msg Message
			err := ws.ReadJSON(&msg)
			if err != nil {
				log.Printf("Hata: %v", err)
				delete(clients, ws)
				break
			}
			broadcast <- msg
		}
	})
	go func() {
		for {
			msg := <-broadcast
			for client := range clients {
				err := client.WriteJSON(msg)
				if err != nil {
					log.Printf("Hata: %v", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}()

	log.Println("Sunucu 5000 portunda çalıştırılıyor")
	err := r.Run(":5000")
	if err != nil {
		log.Fatal(err)
	}
}
