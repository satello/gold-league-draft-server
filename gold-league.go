package main

import (
  "log"
  "net/http"
  "encoding/json"
  "io/ioutil"
  "bytes"
)

// var GOLD_LEAGUE_APP_URI = "https://goldleagueffball.appspot.com"
var GOLD_LEAGUE_APP_URI = "http://localhost:5000"

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
  resp, err := http.Get(GOLD_LEAGUE_APP_URI + "/teams?shuffle=true")
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

  if len(players) < 1 {
    log.Println("Did not fetch players")
    log.Fatal()
  }

  return players
}

func recordBid(player *Player, bidder *Bidder, h *DraftHub) bool {
  log.Println("RECORDING BID")
  values := map[string]interface{}{"name": player.Name, "amount": player.bid.amount, "owner": bidder.Name}
  jsonValue, _ := json.Marshal(values)

  resp, err := http.Post(GOLD_LEAGUE_APP_URI + "/draft/" + h.draftId + "/player/result", "application/json", bytes.NewBuffer(jsonValue))
  if err != nil {
    log.Println(err)
  }

  if resp.StatusCode == 201 {
    return true
  } else {
    return false
  }
}

type RollbackResponse struct {
  Success bool `json:"success"`

  PlayerName string `json:"player"`
}

func rollbackNomination(h *DraftHub) string {
  resp, err := http.Get(GOLD_LEAGUE_APP_URI + "/draft/" + h.draftId + "/rollback-nomination")
  if err != nil {
    log.Println(err)
  }

  if resp.StatusCode != 200 {
    return ""
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Println(err)
  }

  var response *RollbackResponse
  json.Unmarshal(body, &response)

  return response.PlayerName
}
