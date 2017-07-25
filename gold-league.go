package main

import (
  "log"
  "net/http"
  "encoding/json"
  "io/ioutil"
)

var GOLD_LEAGUE_APP_URI = "https://goldleagueffball.appspot.com"

type Owner struct {
  Name string `json:"name"`
  CapRoom int `json:"cap_room"`
  YearsRemaining int `json:"years_remaining"`
  SpotsAvailable int `json:"spots_available"`
}

// cast owner type to bidder
func ownerToBidder(owner *Owner) *Bidder {
  return newBidder(owner.Name, owner.CapRoom, owner.SpotsAvailable)
}


func fetchOwners() []*Owner {
  log.Println("FETCHING OWNERS")
  // fetch owners
  resp, err := http.Get(GOLD_LEAGUE_APP_URI + "/teams")
  if err != nil {
    log.Println(err)
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Println(err)
  }

  var owners []*Owner
  json.Unmarshal(body, &owners)

  if len(owners) != 12 {
    log.Println("Did not fetch 12 owners. Only %d", len(owners))
    log.Fatal()
  }

  return owners
}

func fetchFreeAgents() []*Player {
  log.Println("FETCHING FREE AGENTS")
  // fetch owners
  resp, err := http.Get(GOLD_LEAGUE_APP_URI + "/players/free-agents")
  if err != nil {
    log.Println(err)
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Println(err)
  }

  var players []*Player
  json.Unmarshal(body, &players)
  log.Printf("Number of players in pool: %d", len(players))

  if len(players) < 1 {
    log.Println("Did not fetch players")
    log.Fatal()
  }

  return players
}
