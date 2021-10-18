package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"

	_ "github.com/ClickHouse/clickhouse-go"

	"github.com/google/go-github/github"
	"github.com/jmoiron/sqlx"
)

const MINSTARSLASTYEAR = 1000

type Period struct {
	name string
	days int
}

type DataTable map[string]TableItem

func (d DataTable) Keys() []string {
	keys := make([]string, 0, len(d))

	for k := range d {
		keys = append(keys, k)
	}
	return keys
}

type TableItem struct {
	Growth30  float64
	Growth180 float64
	Growth365 float64
}

func getRepos(connect *sqlx.DB) DataTable {
	var items []string

	selector := `
	SELECT
		repo_name
	FROM github_events
	WHERE event_type = 'WatchEvent' AND created_at > minus(now(), toIntervalDay(365))
	GROUP BY repo_name
	HAVING (count() >= ?);
	`

	if err := connect.Select(&items, selector, MINSTARSLASTYEAR); err != nil {
		log.Fatal(err)
	}

	data := DataTable{}
	for _, item := range items {
		log.Printf("name: %s", item)
		data[item] = TableItem{}
	}
	log.Printf("# items: %d", len(items))
	return data
}

func getGrowths(connect *sqlx.DB, data DataTable) {
	periods := []Period{
		{"1y", 365},
		{"6mo", 180},
		{"1mo", 30},
	}
	selector := `
	WITH dateDiff('day', created_at, (SELECT max(created_at) FROM github_events)) as days
	SELECT
		repo_name,
		sum(days > ?) as penult,
		sum(days <= ?) as ult,
		round((penult + ult) / penult, 3) as growth
	FROM github_events
	WHERE event_type = 'WatchEvent'
	GROUP BY repo_name
	HAVING (max(days) >= ? * 2) and penult > 0 and repo_name in (?)
	ORDER BY growth desc
	`
	for _, period := range periods {
		var items []struct {
			RepoName string  `db:"repo_name"`
			Penult   int32   `db:"penult"`
			Ult      int32   `db:"ult"`
			Growth   float64 `db:"growth"`
		}

		if err := connect.Select(&items, selector, period.days, period.days, period.days, data.Keys()); err != nil {
			log.Fatal(err)
		}

		for _, item := range items {
			log.Printf("name: %s, growth: %f", item.RepoName, item.Growth)
			itemHere := data[item.RepoName]
			switch period.days {
			case 30:
				itemHere.Growth30 = item.Growth
			case 180:
				itemHere.Growth180 = item.Growth
			case 365:
				itemHere.Growth365 = item.Growth
			}

		}
		log.Printf("# items: %d", len(items))

	}
}
func WriteToCSV(d DataTable, jsonMap map[string]RepoInfo) {
	outFile, _ := os.Create("data/out.csv")
	w := csv.NewWriter(outFile)

	if err := w.Write([]string{"name", "url", "stars", "growth30", "growth180", "growth365", "language", "topics", "description"}); err != nil {
		log.Fatalln("OMFG", err)
	}
	for repoName, tableItem := range d {
		var stars, language, topics, description string
		if repoInfo, ok := jsonMap[repoName]; ok {
			var topicsList []string
			stars, language, topicsList, description = fmt.Sprint(repoInfo.Stars), repoInfo.Language, repoInfo.Topics, repoInfo.Description
			topics = strings.Join(topicsList, ", ")
		}

		record := []string{
			repoName,
			"https://github.com/" + repoName,
			stars,
			fmt.Sprint(tableItem.Growth30),
			fmt.Sprint(tableItem.Growth180),
			fmt.Sprint(tableItem.Growth365),
			language,
			topics,
			description,
		}
		fmt.Println("record", record)
		if err := w.Write(record); err != nil {
			log.Println("CSV write error", err)
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}

func WriteToJSON(d DataTable, jsonMap map[string]RepoInfo) {
	type JSONOutItem struct {
		Name        string
		Stars       int
		Growth30    float64
		Growth180   float64
		Growth365   float64
		Language    string
		Topics      string
		Description string
	}
	outItems := []JSONOutItem{}
	for repoName, tableItem := range d {
		var language, topics, description string
		var stars int
		if repoInfo, ok := jsonMap[repoName]; ok {
			var topicsList []string
			stars, language, topicsList, description = repoInfo.Stars, repoInfo.Language, repoInfo.Topics, repoInfo.Description
			topics = strings.Join(topicsList, ", ")
		}

		outItems = append(outItems, JSONOutItem{
			repoName,
			stars,
			tableItem.Growth30,
			tableItem.Growth180,
			tableItem.Growth365,
			language,
			topics,
			description,
		})
	}
	bytes, err := json.Marshal(outItems)
	if err != nil {
		log.Println("Error marshaling JSON", err)
	}
	err = ioutil.WriteFile("data/out.json", bytes, 0644)
	if err != nil {
		log.Println("Error saving JSON", err)
	}

}

type RepoInfo struct {
	Stars       int
	Language    string
	Topics      []string
	Description string
}

const JSONFILENAME = "data/repo-info.json"

func loadJSONMap() map[string]RepoInfo {
	data := make(map[string]RepoInfo)

	jsonBytes, err := ioutil.ReadFile(JSONFILENAME)
	if err != nil {
		log.Println("Error loading JSON", err)
	} else {
		json.Unmarshal(jsonBytes, &data)
	}
	return data
}

func writeJSONMap(data map[string]RepoInfo) {
	bytes, _ := json.Marshal(data)
	if err := ioutil.WriteFile(JSONFILENAME, bytes, 0644); err != nil {
		log.Println(err)
	}
}

func GH(repoNames []string) map[string]RepoInfo {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "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"},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	jsonMap := loadJSONMap()
	log.Println("Loaded JSON Map", jsonMap)
	for _, repoName := range repoNames {
		if _, ok := jsonMap[repoName]; ok {
			// We have this info skip it
			log.Printf("Skipping %s", repoName)
			continue
		}
		log.Println("Getting %s", repoName)
		spluts := strings.Split(repoName, "/")
		owner, name := spluts[0], spluts[1]
		var repo *github.Repository
		var err error
		for {
			repo, _, err = client.Repositories.Get(ctx, owner, name)
			if _, ok := err.(*github.RateLimitError); ok {
				log.Println("Hit rate limit, sleeping 1 minute")
				time.Sleep(time.Minute)
			} else {
				break
			}
		}
		if repo == nil {
			log.Println("Repo is nil:", repoName)
			continue
		}
		stars := 0
		if repo.StargazersCount != nil {
			stars = *repo.StargazersCount
		}
		language := ""
		if repo.Language != nil {
			language = *repo.Language
		}
		description := ""
		if repo.Description != nil {
			description = *repo.Description
		}
		ri := RepoInfo{
			stars,
			language,
			repo.Topics,
			description,
		}
		jsonMap[repoName] = ri
	}
	writeJSONMap(jsonMap)
	return jsonMap
}

func main() {
	connect, err := sqlx.Open("clickhouse", "tcp://gh-api.clickhouse.tech:9440?debug=true&username=explorer&secure=true")
	if err != nil {
		log.Fatal(err)
	}

	// Get repos we want to have
	data := getRepos(connect)
	jsonMap := GH([]string{"dodafin/struba", "facebook/react"})
	// Get GitHub data for these repos (either cached or anew)
	// jsonMap := GH(data.Keys())
	// Get Growth data from ClickHouse
	getGrowths(connect, data)
	log.Println(data)
	// Write out
	// WriteToCSV(data, jsonMap)
	WriteToJSON(data, jsonMap)
}
