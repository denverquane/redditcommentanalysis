# redditcommentanalysis
This is a Go application for injesting and analyzing comments from https://www.reddit.com/r/datasets/comments/3bxlg7/i_have_every_publicly_available_reddit_comment/

## Installation
`go get -d github.com/denverquane/redditcommentanalysis`

`cd $GOROOT/src/github.com/denverquane/redditcommentanalysis`

`go build main.go`

`./main.exe` or `./main`

## Goals
This program is primarily to be used for 

1. Ingesting reddit comment data and extracting important characteristics and properties of the comments, but also 

2. performing statistical analysis and summarizing properties about Reddit's userbase, subreddits, and community tendencies.
