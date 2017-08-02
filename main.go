package main

import (
	"fmt"
	"flag"
	"log"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"os/exec"
	"github.com/rs/cors"
	"strings"
	"time"
)

var addr = flag.String("addr", ":6565", "http service address")

func createUuid() string {
  // return UUID
  out, err := exec.Command("uuidgen").Output()
  if err != nil {
      log.Fatal(err)
  }
  return strings.Replace(string(out[:]), "\n", "", -1)
}

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    log.Printf("%s took %s", name, elapsed)
}

func main() {
	flag.Parse()

	var trumpTower *Tower
	trumpTower = newTower()

	// show draft creator page
	router := httprouter.New()
	// must POST to this route with rules, bidders and players
	router.POST("/new-room", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// hit gold league app to get FA, and owners
		owners := fetchOwners()
		players := fetchFreeAgents()

		// only take 12 bidders
		var bidders []*Bidder
		// convert owners to bidders
		for _, v := range owners {
			b := ownerToBidder(v)
			bidders = append(bidders, b)
		}
		//
		roomId := newRoom(trumpTower, bidders, players)
		// // return room id
		w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, `{"result":"%s"}`, roomId)
	})
	router.GET("/rooms/:roomNumber/connect", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// see if draft room exists
		log.Println("trying to connect....")
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
	router.GET("/rooms", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		json := trumpTower.getRoomsJson()
		w.Write(json)
	})
	router.GET("/ping", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		time.Sleep(time.Second)
		response := Response{"PING", map[string]interface{}{"time": time.Now().UnixNano()}}
		json := responseToJson(response)
		w.Write(json)
	})

	handler := cors.Default().Handler(router)
	err := http.ListenAndServe(*addr, handler)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
