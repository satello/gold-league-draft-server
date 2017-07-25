package main

import (
  "fmt"
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

  // end bidding on player
  endBidding chan *Player

  // players eligable for draft. name to player
  players *PlayersIndex

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

  // bidding loop
  biddingCycle *BiddingCycle
}

func newDraft(bidders []*Bidder, players []*Player) *DraftHub {
  bidder_map := make(map[string]*Bidder)
  for _, v := range bidders {
    bidder_map[v.BidderId] = v
  }

  playersIndex := newPlayersIndex(players)

  nominationCycle := newNominationCycle()
  biddingCycle := newBiddingCycle()

	return &DraftHub{
		broadcast:        make(chan []byte),
		register:         make(chan *Subscriber),
		unregister:       make(chan *Subscriber),
    acceptMessage:    make(chan *Message),
    startBidding:     make(chan *Player),
    endBidding:       make(chan *Player),
		clients:          make(map[*Subscriber]bool),
    curBidderIndex:   0,
    players:          playersIndex,
    biddersMap:       bidder_map,
    biddersSlice:     bidders,
    isActive:         false,
    nominationCycle:  nominationCycle,
    biddingCycle:     biddingCycle,
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

      broadcastNewPlayerNominee(player, h)

      // start bidding cycle
      go h.biddingCycle.getBids(player, h)


    case player := <-h.endBidding:
      log.Println("BIDDING ENDED")
      log.Println(player)

      // subtract cap and space from bidder
      bidder := h.biddersMap[player.bid.bidderId]
      bidder.Cap -= player.bid.amount
      bidder.Spots -= 1

      if bidder.Cap < 1 || bidder.Spots < 1 {
        // mark bidder as unable to draft any longer
        bidder.Draftable = false
      }

      // send out message adjusting bidders cap, spots and eligability
      broadcastBidderState(bidder, h)

      // TODO send something to gold league app to record result

      // make it so player cannot be nominated or bid upon again
      player.taken = true

      // Keep the train rolling
      nextNomination(h)


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
          nextNomination(h)
        }

      case "nextNomination":
        nextNomination(h)


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

        player := h.players.getPlayerByName(playerName)
        if player == nil {
          log.Printf("Shit the bed. %s not in hub", playerName)
          continue
        }
        if player.taken {
          log.Println("Player already bid on")
          continue
        }

        h.nominationCycle.nominationChan <- &Nomination{
          player: player,
          bidderId: bidderId,
        }

      case "bid":
        log.Println("PLACING BID")
        if !h.biddingCycle.open {
          log.Println("biddingCycle isn't open")
          continue
        }
        typed, _ := typed.Json(messageJson.rawJson)

        body := typed.Object("body")
        amount := body.Int("amount")
        bidderId := body.String("bidderId")

        // Check to make sure bid is valid
        if h.biddersMap[bidderId].Cap < amount || h.biddersMap[bidderId].Spots < 1 {
          log.Printf("Bidder %s has insufficient resources to make bid", bidderId)
          continue
        }
        log.Println("trying to give to chan")
        h.biddingCycle.biddingChan <- &Bid{
          amount: amount,
          bidderId: bidderId,
        }

    	case "chatMessage":
        log.Printf("CHAT MESSAGE");
        body := messageJson.Body

        response := Response{"CHAT_MESSAGE", body}
        response_json := responseToJson(response)
        broadcastMessage(h, response_json)

    	default:
    		// freebsd, openbsd,
    		// plan9, windows...
    		fmt.Printf("%s.", t)
      }
		}
	}
}
