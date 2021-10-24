<img src="https://github.com/dodafin/octotrends-frontend/blob/main/src/images/octotrends-logo-black.png?raw=true" width="350" alt="OctoTrends logo">

## [octotrends.com](https://octotrends.com/)

A niftly little tool I wrote to try and find repos and languages that are rapidly growing on GitHub. Growth rates are based on % growth in stars over the past 30 / 180 / 365 days which isn't perfect but it is a way to dig deeper than what [GitHub's Explore](https://github.com/explore) offers.


#### Methodology:

1.  Get a list of repos that've received a certain minimum amount of stars in the past year (200 is the default). 
1.  Aggregate baselines & stars added for 3 time periods. E.g. for the 1 year period, query for stars added in the last year (last period) and stars added in the year before that (baseline period). This allows us to calculate growth percentages.
2.  For each repo in our list, scrape the GitHub API for additional information (current # of stars, primary programming language, description). [GH Archive](https://www.gharchive.org/) doesn't track star removals, and so star counts computed solely on it are inaccurate.
3.  Join the data and write it out to JSON for the [frontend](https://github.com/dodafin/octotrends-frontend) to consume.


Big thanks to the awesome [GH Archive](https://www.gharchive.org/) hosted by [ClickHouse](https://ghe.clickhouse.tech/) which serves as the inspiration and enabler for this project. ❤️