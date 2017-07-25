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
	taken bool
}

func newPlayer(name string, position string) *Player {
  return &Player{
    Name:     name,
    Position: position,
    bid:      &Bid{
			amount: 0,
		},
		taken: false,
  }
}

// func (p *Player) submitBid(bid int, bidder Bidder) bool {
//   if bid > p.Bid && p.bidder != bidder {
//     p.Bid = bid
//     p.bidder = bidder
//     return true
//   } else {
//     return false
//   }
// }

func getPlayers(s *Subscriber, h *DraftHub) {
  log.Printf("GET PLAYERS")
	// return the sorted slice of players
  response := Response{"GET_PLAYERS", map[string]interface{}{"players": h.players.valueSlice}}
  response_json := responseToJson(response)
  sendMessageToSubscriber(h, s, response_json)
}

type PlayersIndex struct {
	// map of name -> player
	nameMap map[string]*Player

	// sorted slice of players from lowest value to highest
	valueSlice []*Player
}

func newPlayersIndex(players []*Player) *PlayersIndex {
	log.Println("making new thing")
	playerMap := make(map[string]*Player)
	for _, v := range players {
		playerMap[v.Name] = v
	}
	log.Println("slice is gucci")

	// sort players by value
	log.Println(len(players))
	sort.Slice(players, func(i, j int) bool { return players[i].Value < players[j].Value })
	log.Printf("length of players sorted slice %d", len(players))
	log.Println(players[len(players)-1])

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
		if !player.taken {
			break
		}
		lastIndex -= 1
	}
	return player
}
