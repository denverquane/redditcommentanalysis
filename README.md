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

## Usage
The program relies on several values found in the .env file at the project root (a sample.env is provided with this repo).
For example, enabling the program to run as a server, responding to various job invocation requests, or providing the status
of a particular data processing task. This is also where the base data directory is defined, and the structure of this directory is
very important.

Structure:

    BASEDIR
    -> RC_2015_01
    -> RC_2016_01
    -> RC_2016_02
    -> ...
    -> RC_2017_03

The program will automatically generate "Extracted" and "Processed" directories to store the results of extracting and processing files,
respectively

RC_<year>_<month> files indicate the extracted data from https://www.reddit.com/r/datasets/comments/3bxlg7/i_have_every_publicly_available_reddit_comment/
for each respective month/year
