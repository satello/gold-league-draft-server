package main

import (
	"fmt"
	"flag"
	"log"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"os/exec"
	"encoding/json"
	"github.com/rs/cors"
	"strings"
)

var addr = flag.String("addr", ":5000", "http service address")

func createUuid() string {
  // return UUID
  out, err := exec.Command("uuidgen").Output()
  if err != nil {
      log.Fatal(err)
  }
  return strings.Replace(string(out[:]), "\n", "", -1)
}

func serveHome(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}

	http.ServeFile(w, r, "html/index.html")
}

func main() {
	flag.Parse()

	var trumpTower *Tower
	trumpTower = newTower()

	// show draft creator page
	router := httprouter.New()
	router.GET("/", serveHome)
	// must POST to this route with rules, bidders and players
	router.POST("/new-room", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// parse JSON body
		var body Room
    if r.Body == nil {
        http.Error(w, "Please send a request body", 400)
        return
    }
    err := json.NewDecoder(r.Body).Decode(&body)
    if err != nil {
        http.Error(w, err.Error(), 400)
        return
    }

		// assign bidder ids
		// TODO should it really be here?
		for _, v := range body.Bidders {
			new_uuid := createUuid()
			v.BidderId = new_uuid
		}

		roomId := newRoom(trumpTower, body.Rules, body.Bidders, body.Players)
		// return room id
		w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, `{"result":"%s"}`, roomId)
	})
	router.GET("/:roomNumber/connect", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// see if draft room exists
		log.Println("trying to connect....")
		log.Println(len(trumpTower.rooms))
		log.Println(ps.ByName("roomNumber"))
		hub := trumpTower.rooms[ps.ByName("roomNumber")]
		if hub != nil {
			serveWs(hub, w, r)
		} else {
			log.Println("draft room does not exist")
			http.Error(w, "Draft Room does not exist", 401)
			return
		}
	})

	handler := cors.Default().Handler(router)
	err := http.ListenAndServe(*addr, handler)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
