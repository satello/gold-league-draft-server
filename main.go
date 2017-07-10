package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":5000", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	http.ServeFile(w, r, "html/index.html")
}

func main() {
	flag.Parse()

	hub := newDraft()
	go hub.run()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
