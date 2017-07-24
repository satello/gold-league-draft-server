package main

import (
  "fmt"
  "encoding/json"
  "log"
  "github.com/mitchellh/mapstructure"
  "github.com/karlseguin/typed"
)

// draft hub maintains the set of active clients and broadcasts messages to the
// clients.
type DraftHub struct {
	// Registered clients.
	clients map[*Subscriber]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Subscriber

	// Unregister requests from clients.
	unregister chan *Subscriber

  // accept message from client
  acceptMessage chan *Message

  // channel for nomination cycle to communicate with hub
  startBidding chan *Player

  // players eligable for draft. name to player
  players map[string]*Player

  // store order of bidders
  biddersSlice []*Bidder

  // bidders in the draft
  biddersMap map[string]*Bidder

  // current bidder index
  curBidderIndex int

  // flag to set when you want to close draft room
  isActive bool

  // -- LOOPS --
  // nomination loop
  nominationCycle *NominationCycle
}

func newDraft(bidders []*Bidder, players []*Player) *DraftHub {
  bidder_map := make(map[string]*Bidder)
  for _, v := range bidders {
    bidder_map[v.BidderId] = v
  }

  player_map := make(map[string]*Player)
  for _, v := range players {
    player_map[v.Name] = v
  }

  nominationCycle := newNominationCycle()

	return &DraftHub{
		broadcast:        make(chan []byte),
		register:         make(chan *Subscriber),
		unregister:       make(chan *Subscriber),
    acceptMessage:    make(chan *Message),
    startBidding:     make(chan *Player),
		clients:          make(map[*Subscriber]bool),
    curBidderIndex:   0,
    players:          player_map,
    biddersMap:       bidder_map,
    biddersSlice:     bidders,
    isActive:         false,
    nominationCycle:  nominationCycle,
	}
}

func (h *DraftHub) run() {
  // handle draft related tasks
	for {
		select {

		case client := <-h.register:
      log.Println("CONNECTING CLIENT")
			h.clients[client] = true

		case client := <-h.unregister:
      log.Println("DISCONNECTING CLIENT")
			if _, ok := h.clients[client]; ok {
        // mark bidder as inactive
        b := h.biddersMap[client.bidderId]
        if b != nil {
          b.ActiveConnection = false
          broadcastBidderState(b, h)
        }
        // remove client
				delete(h.clients, client)
        // close clinet
				close(client.send)
			}

    case player := <-h.startBidding:
      h.curBidderIndex += 1
      log.Println("GLORIOUS DAY")
      log.Println(player)

		case messageJson := <-h.acceptMessage:
      switch t := messageJson.MessageType; t {

      case "authorizeBidder":
        var body TokenBody
        mapstructure.Decode(messageJson.Body, &body)

        token := body.Token
        authorizeBidder(token, messageJson.Subscriber, h)

      case "deauthorizeBidder":
        var body TokenBody
        mapstructure.Decode(messageJson.Body, &body)

        token := body.Token
        deactivateBidder(token, messageJson.Subscriber, h)

      case "getBidders":
        getBidders(messageJson.Subscriber, h)

      case "getPlayers":
        getPlayers(messageJson.Subscriber, h)

      case "startDraft":
        if !h.isActive {
          h.isActive = true
          firstBidder := h.biddersSlice[h.curBidderIndex]
          // send to front end who is allowed to make first nomination
          broadcastNewBidderNominee(firstBidder, h)

          // start the clock
          go h.nominationCycle.getNominee(h)
        }

      case "nextNomination":
        firstBidder := h.biddersSlice[h.curBidderIndex]
        // send to front end who is allowed to make first nomination
        broadcastNewBidderNominee(firstBidder, h)

        // start the clock
        go h.nominationCycle.getNominee(h)

      case "nominatePlayer":
        log.Println("NOMINATING PLAYER")
        if !h.nominationCycle.open {
          log.Println("nominationCycle isn't open")
          continue
        }
        typed, _ := typed.Json(messageJson.rawJson)

        body := typed.Object("body")
        playerName := body.String("name")
        bidderId := body.String("bidderId")
        if bidderId != h.biddersSlice[h.curBidderIndex].BidderId {
          log.Println("BAD NOMINATOR")
          continue
        }
        log.Println(playerName)
        player := h.players[playerName]
        if player == nil {
          log.Printf("Shit the bed. %s not in hub", playerName)
        }

        log.Println("trying to nominate")
        h.nominationCycle.nominationChan <- &Nomination{
          player: player,
          bidderId: bidderId,
        }

    	case "chatMessage":
        log.Printf("CHAT MESSAGE");
        body := messageJson.Body

        response := Response{"CHAT_MESSAGE", body}
        response_json, err := json.Marshal(response)
        if err != nil {
    			log.Printf("error: %v", err)
    			break
        }
        broadcastMessage(h, response_json)

    	default:
    		// freebsd, openbsd,
    		// plan9, windows...
    		fmt.Printf("%s.", t)
      }
		}
	}
}
