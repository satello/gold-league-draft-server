package main

type Response struct {
  MessageType string
  Body map[string]interface{} `json:"body"`
}
