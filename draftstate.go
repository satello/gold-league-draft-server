package main


type DraftState struct {
  // bidder id of the current bidder with highest bid
  CurrentBidderId string `json:"currentBidderId"`

  // current player up for bid
  CurrentPlayerName string `json:"currentPlayerName"`

  // current highest bid
  CurrentBid int `json:"currentBid"`

  // current nominator
  CurrentNominatorId string `json:"currentNominatorId"`

  // if draft is active
  Running bool `json:"draftRunning"`

  // if draft is paused
  Paused bool `json:"paused"`

  // if we are currently nominating
  nominating bool

  // if we are currently bidding
  bidding bool
}

func sendDraftState(s *Subscriber, h *DraftHub) {
  response := Response{"INIT_DRAFT_STATE", map[string]interface{}{"draftState": h.draftState}}
  response_json := responseToJson(response)
  sendMessageToSubscriber(h, s, response_json)
}
