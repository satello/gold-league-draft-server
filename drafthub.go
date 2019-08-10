package main

import (
  "log"
  "github.com/mitchellh/mapstructure"
  "github.com/karlseguin/typed"
  "time"
)

// draft hub maintains the set of active clients and broadcasts messages to the
// clients.
type DraftHub struct {
  // draft identifying string
  draftId string

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

  // channel to begin rollback of nomination
  rollbackChan chan bool

  // chan to hit when the draft is over
  endDraftChan chan bool

  // chan for starting next nomination
  nextNominationChan chan bool

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

  // state of the draft. to help frontend bootstrap
  draftState *DraftState
}

func newDraft(bidders []*Bidder, players []*Player, roomId string) *DraftHub {
  bidder_map := make(map[string]*Bidder)
  for _, v := range bidders {
    bidder_map[v.BidderId] = v
  }

  playersIndex := newPlayersIndex(players)

  nominationCycle := newNominationCycle()
  biddingCycle := newBiddingCycle()

	return &DraftHub{
    draftId:            roomId,
		broadcast:          make(chan []byte),
		register:           make(chan *Subscriber),
		unregister:         make(chan *Subscriber),
    acceptMessage:      make(chan *Message),
    startBidding:       make(chan *Player),
    endBidding:         make(chan *Player),
    rollbackChan:       make(chan bool),
    endDraftChan:       make(chan bool),
    nextNominationChan: make(chan bool),
		clients:            make(map[*Subscriber]bool),
    curBidderIndex:     0,
    players:            playersIndex,
    biddersMap:         bidder_map,
    biddersSlice:       bidders,
    isActive:           false,
    nominationCycle:    nominationCycle,
    biddingCycle:       biddingCycle,
    draftState:         &DraftState{},
	}
}

func (h *DraftHub) run() {
  // handle draft related tasks
  mainLoop:
	for {
		select {

		case client := <-h.register:
      log.Println("CONNECTING CLIENT")
			h.clients[client] = true
      // send the current state of the draft to client
      sendDraftState(client, h)

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
      log.Println("STARTING BIDDING")
      h.curBidderIndex = (h.curBidderIndex + 1) % len(h.biddersSlice)
      h.draftState.nominating = false

      // braodcast nominee to clients
      broadcastNewPlayerNominee(player, h)

      // start bidding cycle
      h.draftState.bidding = true
      go h.biddingCycle.getBids(player, h)

    case player := <-h.endBidding:
      log.Println("BIDDING ENDED")

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

      // Record bid via flask service
      // TODO just write to file ourselves?
      success := recordBid(player, bidder, h)
      if !success {
        // if we can't record bid try 5 more times and then blow up
        go func() {
          retries := 5
          backoff := 1
          for {
            if retries == 0 {
              break
            }
            success := recordBid(player, bidder, h)
            if success {
              // if we get things recorded then we are good
              return
            }
            // backoff
            time.Sleep(time.Second * time.Duration(backoff))
            // multiply
            backoff *= 2
            retries -= 1
          }
          // crash program if we can't write cause we should really see whats going on
          log.Fatal("UNABLE TO WRITE RESULTS TO FLASK --- CHECK IF FLASK SERVER RUNNING")
        }()
      }

      // make it so player cannot be nominated or bid upon again
      player.Taken = true
      h.draftState.bidding = false
      braodcastPlayers(h)

      // Keep the train rolling
      nextNomination(h)

    case <- h.rollbackChan:
      // make new routine so that drafthub doesn't deadlock
      go previousNomination(h)

    case <- h.nextNominationChan:
      nextNomination(h)

    case <- h.endDraftChan:
      // close any open cycles
      if h.nominationCycle.open {
        h.nominationCycle.interuptChan <- make(chan bool)
      }
      if h.biddingCycle.open {
        h.biddingCycle.interuptChan <- make(chan bool)
      }

      h.draftState.Running = false
      // end drafthub
      // TODO do we want to do this? or leave open so subscribers can still connect?
      break mainLoop

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
          h.draftState.Running = true
          h.draftState.Paused = false
          nextNomination(h)
        }

      case "nextNomination":
        nextNomination(h)

      case "rollbackNomination":
        log.Println("ROLLING BACK NOMINATION")
        if h.draftState.Running == false {
          continue
        }
        playerName := rollbackNomination(h)
        if playerName != "" {
          // kind of hacky but since we can't pass player to rollbackChan we just save the state
          h.draftState.CurrentPlayerName = playerName
          if h.nominationCycle.open {
            log.Println("stopping nomination cycle...")
            h.nominationCycle.interuptChan <- h.rollbackChan
          } else if h.biddingCycle.open {
            log.Println("stopping bidding cycle...")
            h.biddingCycle.interuptChan <- h.rollbackChan
            // have to take off cur bidder index
            if (h.curBidderIndex == 0) {
              h.curBidderIndex = 12
            }

            h.curBidderIndex -= 1
          }
        } else {
          continue
        }

      case "nominatePlayer":
        log.Println("NOMINATING PLAYER")
        if !h.nominationCycle.open {
          log.Println("nominationCycle isn't open")
          continue
        }
        // easier way to handle json
        // TODO do this for all messages
        typed, _ := typed.Json(messageJson.rawJson)
        body := typed.Object("body")
        playerName := body.String("name")
        amount := body.Int("amount")
        bidderId := body.String("bidderId")
        // see if bidder is supposed to be making nomination
        if bidderId != h.biddersSlice[h.curBidderIndex].BidderId {
          log.Println("BAD NOMINATOR")
          continue
        }
        // make sure player eligible
        player := h.players.getPlayerByName(playerName)
        if player == nil {
          log.Printf("Shit the bed. player %s not in hub", playerName)
          continue
        }
        if player.Taken {
          log.Println("Player already bid on")
          continue
        }
        // make sure bidder eligible
        bidder := h.biddersMap[bidderId]
        if bidder == nil {
          log.Printf("Shit the bed. bidder %s not in hub", bidder.Name)
          continue
        }
        if !bidder.Draftable {
          log.Println("Bidder who is not draftable cannot make nomination")
          continue
        }
        if bidder.Cap < amount {
          log.Println("Bidder does not have enough funds to make bid")
          continue
        }
        // place nomination
        h.nominationCycle.nominationChan <- &Nomination{
          player: player,
          bidderId: bidderId,
          amount: amount,
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

        // make sure bidder eligible
        bidder := h.biddersMap[bidderId]
        if bidder == nil {
          log.Printf("Cannot place bid: Bidder %s is not in drafthub", bidderId)
          continue
        }
        // make sure bidder is still active
        if !bidder.Draftable {
          log.Printf("Cannot place bid: Bidder %s has already left draft", bidderId)
          continue
        }

        // Check to make sure bid is valid
        if bidder.Cap < amount || bidder.Spots < 1 {
          log.Printf("Cannot place bid: Bidder %s has insufficient resources", bidderId)
          continue
        }
        // place bid
        h.biddingCycle.biddingChan <- &Bid{
          amount: amount,
          bidderId: bidderId,
        }

      case "pauseDraft":
        log.Printf("PAUSING DRAFT")
        if h.draftState.nominating {
          h.nominationCycle.pauseChan <- true
        } else if h.draftState.bidding {
          h.biddingCycle.pauseChan <- true
        }
        h.draftState.Paused = true
        response := Response{"DRAFT_PAUSED", nil}
        response_json := responseToJson(response)
        broadcastMessage(h, response_json)

      case "resumeDraft":
        log.Printf("RESUMING DRAFT")
        if h.draftState.nominating {
          h.nominationCycle.pauseChan <- false
        } else if h.draftState.bidding {
          h.biddingCycle.pauseChan <- false
        }
        h.draftState.Paused = false
        response := Response{"DRAFT_RESUMED", nil}
        response_json := responseToJson(response)
        broadcastMessage(h, response_json)

      case "withdrawBidder":
        typed, _ := typed.Json(messageJson.rawJson)

        body := typed.Object("body")
        bidderId := body.String("bidderId")
        log.Printf("BIDDER %s LEAVING DRAFT", bidderId)

        bidder := h.biddersMap[bidderId]
        if bidder == nil {
          log.Printf("Cannot Leave: bidder %s does not exist in hub", bidderId)
          continue
        }
        if bidder.Draftable == false {
          continue
        }
        // cannot leave if you currently have the highest bid
        if h.draftState.CurrentBidderId == bidderId && h.draftState.bidding {
          log.Printf("Cannot Leave: bidder is highest bidder", bidderId)
          continue
        }
        // if nominator mark as not draftable and find next nominator
        if h.draftState.CurrentNominatorId == bidderId && h.draftState.nominating {
          bidder.Draftable = false
          broadcastBidderState(bidder, h)
          h.nominationCycle.interuptChan <- h.nextNominationChan
        } else {
          bidder.Draftable = false
          broadcastBidderState(bidder, h)

        }

    	case "chatMessage":
        log.Printf("CHAT MESSAGE")
        body := messageJson.Body

        response := Response{"CHAT_MESSAGE", body}
        response_json := responseToJson(response)
        broadcastMessage(h, response_json)

    	default:
    		// freebsd, openbsd,
    		// plan9, windows...
    		log.Printf("%s.", t)
      }
		}
	}
}
