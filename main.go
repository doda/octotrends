package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go"

	"github.com/jmoiron/sqlx"
)

const MINSTARSLASTYEAR = 10000

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
		// log.Printf("name: %s", item)
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
			// log.Printf("name: %s, growth: %f", item.RepoName, item.Growth)
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

func main() {
	connect, err := sqlx.Open("clickhouse", "tcp://gh-api.clickhouse.tech:9440?debug=false&username=explorer&secure=true")
	if err != nil {
		log.Fatal(err)
	}

	// Get repos we want to have
	data := getRepos(connect)
	// Get GitHub data for these repos (either cached or anew)
	jsonMap := GH([]string{"dodafin/struba", "facebook/react", "doesnotexist/doesnotexist2000"})
	// jsonMap := GH(data.Keys())
	// Get Growth data from ClickHouse
	getGrowths(connect, data)
	// log.Println(data)
	// Write out
	WriteToJSON(data, jsonMap)
}
