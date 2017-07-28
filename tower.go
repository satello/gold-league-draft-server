package main

import (
  "log"
  "time"
  "encoding/json"
)

type Tower struct {
  rooms map[string]*DraftHub
}

func newTower() *Tower {
  return &Tower{
    rooms: make(map[string]*DraftHub),
  }
}

// type Room struct {
//   RoomId string `json:"roomId"`
//
//   Bidders []*Bidder `json:"bidders"`
//
//   Players []*Player `json:"players"`
// }

func newRoom(t *Tower, bidders []*Bidder, players []*Player) string {
  roomId := createUuid()
  newDraftRoom := newDraft(bidders, players, roomId)

  // start new hub
  go newDraftRoom.run()
  timer := time.NewTimer(time.Hour * 24)
  go func() {
    <- timer.C
    newDraftRoom.isActive = false
    newDraftRoom.draftState.Running = false
    delete(t.rooms, roomId)
    log.Printf("room %s no longer active", roomId)
  }()

  // TODO watch out or memory leaks with this. Do go routines shut down when the parent does?
  t.rooms[roomId] = newDraftRoom

  return roomId
}

type RoomInfo struct {
  RoomId string `json:"roomId"`

  IsRunning bool `json:"isRunning"`

  ActiveConnections int `json:"activeConnections"`

}

// get all rooms
func (t *Tower) getRoomsJson() []byte {
  result := make([]*RoomInfo, len(t.rooms))
  count := 0
  for _, h := range t.rooms {
    result[count] = &RoomInfo{
      RoomId: h.draftId,
      IsRunning: h.draftState.Running,
      ActiveConnections: len(h.clients),
    }
    count += 1
  }

  response_json, err := json.Marshal(result)
  if err != nil {
    log.Printf("error: %v", err)
  }
  return response_json
}
