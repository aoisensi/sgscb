package main

import (
	"encoding/json"
	"os"
)

var config *Config

type Config struct {
	Actors      []*Actor `json:"actors"`
	SteamAPIKey string   `json:"steam_api_key"`
}

type Actor struct {
	Handle   string  `json:"handle"`
	Password string  `json:"password"`
	Appid    string  `json:"appid"`
	Stats    []*Stat `json:"stats"`
}

type Stat struct {
	Name    string `json:"name"`
	Display string `json:"display"`
}

func init() {
	f, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		panic(err)
	}
}
