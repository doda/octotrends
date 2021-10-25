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
	Baseline30  int
	Added30     int
	Baseline180 int
	Added180    int
	Baseline365 int
	Added365    int
}

type JSONOutItem struct {
	Name        string
	Stars       int
	Baseline30  int
	Added30     int
	Baseline180 int
	Added180    int
	Baseline365 int
	Added365    int
	Language    string
	Topics      string
	Description string
}

var RepoSelectQuery = `
SELECT
	repo_name
FROM github_events
WHERE event_type = 'WatchEvent' AND created_at > minus(now(), toIntervalDay(365))
GROUP BY repo_name
HAVING (count() >= ?)
`

var StarsSelectQuery = `
WITH dateDiff('day', created_at, (SELECT max(created_at) FROM github_events)) as days
SELECT
	repo_name,
	sum(days > ?) as baseline,
	sum(days <= ?) as added
FROM github_events
WHERE event_type = 'WatchEvent'
GROUP BY repo_name
HAVING repo_name in (` + RepoSelectQuery + `)
`

// Get an initial list of repos to scrape that meet a minimum popularity bar
func GetRepos(connect *sqlx.DB, minStars int) DataTable {
	var items []string

	if err := connect.Select(&items, RepoSelectQuery, minStars); err != nil {
		log.Fatal(err)
	}

	data := DataTable{}
	for _, repoName := range items {
		data[repoName] = TableItem{}
	}
	log.Printf("# repos to scrape: %d", len(items))
	return data
}

// Use ClickHouse to quickly aggregate how many people have starred the repo in the "baseline" (second-to-last) period and the
// "added" (last) period. This allows us to calculate the % growth within the "added" period.
func GetGrowths(connect *sqlx.DB, data DataTable, minStars int) {
	periods := []Period{
		{"1mo", 30},
		{"6mo", 180},
		{"1y", 365},
	}
	for _, period := range periods {
		var items []struct {
			RepoName string `db:"repo_name"`
			Baseline int32  `db:"baseline"`
			Added    int32  `db:"added"`
		}
		fmt.Println("Running", StarsSelectQuery)
		if err := connect.Select(&items, StarsSelectQuery, period.days, period.days, minStars); err != nil {
			log.Fatal(err)
		}

		for _, item := range items {
			dataItem := data[item.RepoName]
			base, added := int(item.Baseline), int(item.Added)
			switch period.days {
			case 30:
				dataItem.Baseline30, dataItem.Added30 = base, added
			case 180:
				dataItem.Baseline180, dataItem.Added180 = base, added
			case 365:
				dataItem.Baseline365, dataItem.Added365 = base, added
			}
			data[item.RepoName] = dataItem
		}
		log.Printf("Aggregated %d items for period %d days", len(items), period.days)

	}
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
			Baseline30:  tableItem.Baseline30,
			Added30:     tableItem.Added30,
			Baseline180: tableItem.Baseline180,
			Added180:    tableItem.Added180,
			Baseline365: tableItem.Baseline365,
			Added365:    tableItem.Added365,
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

	log.Println("Getting repos that have received more than", *minStars, "stars in the past year, and using ClickHouse URL:", *clickHouseURL)

	connect, err := sqlx.Open("clickhouse", *clickHouseURL)
	if err != nil {
		log.Fatal(err)
	}

	// Get repos we want to have
	dataTable := GetRepos(connect, *minStars)

	// Get GitHub dataTable for these repos (either cached or anew)
	GHInfoMap := GetGHRepoInfo(dataTable, *githbToken)

	// Get Growth dataTable from ClickHouse
	GetGrowths(connect, dataTable, *minStars)

	// Write out
	WriteToJSON(dataTable, GHInfoMap, *outFileName)
}
