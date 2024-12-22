package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func getGlobalStatsForGame(appid string, names []string) (map[string]int64, error) {
	values := url.Values{
		"key":   {config.SteamAPIKey},
		"appid": {appid},
		"count": {strconv.Itoa(len(names))},
	}
	for i, name := range names {
		values.Add("name["+strconv.Itoa(i)+"]", name)
	}
	url := "https://api.steampowered.com/ISteamUserStats/GetGlobalStatsForGame/v1/?"
	url += values.Encode()
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam api error: %s", resp.Status)
	}
	var data struct {
		Response struct {
			GlobalStats map[string]struct {
				Total int64 `json:"total,string"`
			} `json:"globalstats,omitempty"`
			Result int    `json:"result"`
			Error  string `json:"error,omitempty"`
		} `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if data.Response.Result != 1 {
		return nil, fmt.Errorf("steam api error: %s", data.Response.Error)
	}
	stats := make(map[string]int64, len(names))
	for _, name := range names {
		stats[name] = data.Response.GlobalStats[name].Total
	}
	return stats, nil
}
