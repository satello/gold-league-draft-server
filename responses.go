package main

import (
  "encoding/json"
  "log"
)

type Response struct {
  MessageType string
  Body map[string]interface{} `json:"body"`
}

func responseToJson(message Response) []byte {
  response_json, err := json.Marshal(message)
  if err != nil {
    log.Printf("error: %v", err)
  }
  return response_json
}
