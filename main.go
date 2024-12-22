package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
)

var wg sync.WaitGroup

func main() {
	wg.Add(len(config.Actors))
	for _, actor := range config.Actors {
		go func(actor *Actor) {
			defer wg.Done()
			names := make([]string, len(actor.Stats))
			for i, stat := range actor.Stats {
				names[i] = stat.Name
			}
			latest, err := openLatest(actor.Appid)
			if err != nil {
				log.Printf("open latest failed: %v", err)
				return
			}
			stats, err := getGlobalStatsForGame(actor.Appid, names)
			if err != nil {
				log.Printf("get global stats for game failed: %v", err)
				return
			}
			if latest == nil {
				saveLatest(actor.Appid, stats)
				log.Println("only saved latest")
				return
			}
			texts := make([]string, len(actor.Stats))
			for i, stat := range actor.Stats {
				texts[i] = fmt.Sprintf(
					"%s: %s (%s)",
					stat.Display,
					formatNumber(stats[stat.Name]),
					formatPN(stats[stat.Name]-latest[stat.Name]))
			}
			text := strings.Join(texts, "\n")
			cli, err := login(actor)
			if err != nil {
				log.Printf("login failed: %v", err)
				return
			}
			output, err := post(cli, text)
			if err != nil {
				log.Printf("post failed: %v", err)
				return
			}
			log.Printf("post success: %s", output.Uri)
		}(actor)
	}
	wg.Wait()
}

func login(actor *Actor) (*xrpc.Client, error) {
	cli := &xrpc.Client{
		Host: "https://bsky.social",
	}
	input := &atproto.ServerCreateSession_Input{
		Identifier: actor.Handle,
		Password:   actor.Password,
	}
	output, err := atproto.ServerCreateSession(context.Background(), cli, input)
	if err != nil {
		return nil, err
	}
	cli.Auth = &xrpc.AuthInfo{
		AccessJwt:  output.AccessJwt,
		RefreshJwt: output.RefreshJwt,
		Handle:     output.Handle,
		Did:        output.Did,
	}
	return cli, nil
}

func openLatest(appid string) (map[string]int64, error) {
	f, err := os.Open(fmt.Sprintf("latest_%s.json", appid))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var data map[string]int64
	return data, json.NewDecoder(f).Decode(&data)
}

func saveLatest(appid string, data map[string]int64) error {
	f, err := os.Create(fmt.Sprintf("latest_%s.json", appid))
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(data)
}

func post(cli *xrpc.Client, text string) (*atproto.RepoCreateRecord_Output, error) {
	input := &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.post",
		Repo:       cli.Auth.Did,
		Record: &lexutil.LexiconTypeDecoder{
			Val: &bsky.FeedPost{
				Text:      text,
				CreatedAt: truncateToHour(time.Now()).Format(time.RFC3339),
			},
		},
	}
	return atproto.RepoCreateRecord(context.Background(), cli, input)

}

func truncateToHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

func formatNumber(n int64) string {
	return message.NewPrinter(language.English).Sprintf("%d", n)
}

func formatPN(n int64) string {
	if n == 0 {
		return "±0"
	}
	return fmt.Sprintf("%+d", n)
}
