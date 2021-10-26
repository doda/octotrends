package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/google/go-github/github"

	"github.com/jmoiron/sqlx"
)

type Period struct {
	name string
	days int
}

type DataTable map[string]TableItem

type TableItem struct {
	Added7  int
	Added30 int
	Added90 int
}

type JSONOutItem struct {
	Name        string
	Stars       int
	Added7      int
	Added30     int
	Added90     int
	Language    string
	Topics      string
	Description string
}

var StarsSelectQuery = `
WITH 
	(SELECT max(created_at) FROM github_events) as right_now,
	dateDiff('day', created_at_sub, right_now) as days
SELECT
	repo_name,
	sum(days <= 7) as added7,
	sum(days <= 30) as added30,
	sum(days <= 90) as added90
FROM 
( // at most 1 star per user per repo
	SELECT
		repo_name,
		actor_login,
		any(created_at) as created_at_sub
	FROM github_events
	WHERE event_type = 'WatchEvent'
	GROUP BY repo_name, actor_login
)
WHERE repo_name in 
( // "interesting" repos list
	SELECT
		repo_name
	FROM github_events
	WHERE event_type = 'WatchEvent' AND created_at > minus(right_now, toIntervalDay(90))
	GROUP BY repo_name
	HAVING (count() >= ?)
)
GROUP BY repo_name
`

// Use ClickHouse to quickly aggregate how many people have starred the repo recently
func GetGrowths(connect *sqlx.DB, minStars int) DataTable {
	data := DataTable{}

	var items []struct {
		RepoName string `db:"repo_name"`
		Added7   int32  `db:"added7"`
		Added30  int32  `db:"added30"`
		Added90  int32  `db:"added90"`
	}

	fmt.Println("Running", StarsSelectQuery)
	if err := connect.Select(&items, StarsSelectQuery, minStars); err != nil {
		log.Fatal(err)
	}

	for _, item := range items {
		dataItem := data[item.RepoName]
		dataItem.Added7 = int(item.Added7)
		dataItem.Added30 = int(item.Added30)
		dataItem.Added90 = int(item.Added90)
		data[item.RepoName] = dataItem
	}
	log.Printf("Aggregated %d items", len(items))
	return data
}

// Write our accumulated & joined information out to JSON for the frontend to consume
func WriteToJSON(d DataTable, jsonMap map[string]github.Repository, outFileName string) {
	outItems := []JSONOutItem{}
	for repoName, tableItem := range d {
		repoInfo := jsonMap[repoName]

		language := StringValue(repoInfo.Language)
		if RepoLangBlocked(repoName) {
			language = ""
		}

		outItems = append(outItems, JSONOutItem{
			Name:        repoName,
			Stars:       IntValue(repoInfo.StargazersCount),
			Added7:      tableItem.Added7,
			Added30:     tableItem.Added30,
			Added90:     tableItem.Added90,
			Language:    language,
			Topics:      strings.Join(repoInfo.Topics, ", "),
			Description: StringValue(repoInfo.Description),
		})
	}
	bytes, err := json.Marshal(outItems)
	if err != nil {
		log.Println("Error marshaling JSON", err)
	}
	err = ioutil.WriteFile(outFileName, bytes, 0644)
	if err != nil {
		log.Println("Error saving JSON", err)
	}

}

func main() {
	minStars := flag.Int("minstars", 200, "Minimum stars received in past year to be included")
	clickHouseURL := flag.String("clickhouse", "tcp://gh-api.clickhouse.tech:9440?debug=false&username=explorer&secure=true", "ClickHouse TCP URL")
	githbToken := flag.String("ghp", "", "GitHub Access token")
	outFileName := flag.String("o", "out.json", "Output file name")

	flag.Parse()

	if *githbToken == "" {
		log.Println("Requires a GitHub Access Token")
		return
	}

	log.Printf("Getting repos that have received more than %d stars in the past 90 days, and using ClickHouse URL: %s", *minStars, *clickHouseURL)

	connect, err := sqlx.Open("clickhouse", *clickHouseURL)
	if err != nil {
		log.Fatal(err)
	}

	// Get Growth dataTable from ClickHouse
	dataTable := GetGrowths(connect, *minStars)

	// Get GitHub dataTable for these repos (either cached or anew)
	GHInfoMap := GetGHRepoInfo(dataTable, *githbToken)

	// Write out
	WriteToJSON(dataTable, GHInfoMap, *outFileName)
}
