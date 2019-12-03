package main

import (
	"net/http"
	"log"

)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func servePage(w http.ResponseWriter, r *http.Request) {
	websocket.Upgrader
	if r.URL.Path != "/" {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "public/index.html")
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upg
}

func main() {
	http.HandleFunc("/", FrontPage)

	err := http.ListenAndServe(":8001", nil)

	if err != nil {
		log.Fatal(err)
	}
}
