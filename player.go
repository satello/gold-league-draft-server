package main

import (
	"log"
	"sort"
)

type Player struct {
  // player name
	Name string `json:"name"`

  // player position
	Position string `json:"position"`

	// player bye week
	Bye int `json:"bye"`

	// arbitrary value metric
	Value int `json:"value"`

  // current bid
  bid *Bid

	// if they are already bidded on
	Taken bool `json:"taken"`
}

func newPlayer(name string, position string) *Player {
  return &Player{
    Name:     name,
    Position: position,
    bid:      &Bid{
			amount: 0,
		},
		Taken: false,
  }
}

func getPlayers(s *Subscriber, h *DraftHub) {
  log.Printf("GET PLAYERS")
	// return the sorted slice of players
  response := Response{"GET_PLAYERS", map[string]interface{}{"players": h.players.valueSlice}}
  response_json := responseToJson(response)
  sendMessageToSubscriber(h, s, response_json)
}

func braodcastPlayers(h *DraftHub) {
  log.Printf("BROADCAST PLAYERS")
	// return the sorted slice of players
  response := Response{"GET_PLAYERS", map[string]interface{}{"players": h.players.valueSlice}}
  response_json := responseToJson(response)
  broadcastMessage(h, response_json)
}

type PlayersIndex struct {
	// map of name -> player
	nameMap map[string]*Player

	// sorted slice of players from lowest value to highest
	valueSlice []*Player
}

func newPlayersIndex(players []*Player) *PlayersIndex {
	playerMap := make(map[string]*Player)
	for _, v := range players {
		playerMap[v.Name] = v
	}
	// sort players by value
	sort.Slice(players, func(i, j int) bool { return players[i].Value < players[j].Value })

	return &PlayersIndex{
		nameMap: playerMap,
		valueSlice: players,
	}
}

func (p *PlayersIndex) getPlayerByName(name string) *Player {
	return p.nameMap[name]
}

func (p *PlayersIndex) getHighestValuePlayer() *Player {
	lastIndex := (len(p.valueSlice) - 1)
	var player *Player
	// seems impossible that we would run out of players so not gonna worry about error case of none left for now
	for {
		player = p.valueSlice[lastIndex]
		if !player.Taken {
			break
		}
		lastIndex -= 1
	}
	return player
}
