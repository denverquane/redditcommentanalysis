package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/valyala/fastjson"
	"log"
	"os"
	"strconv"
	"time"
)

type BasicCommentData struct {
	author string
	subreddit string
	body   string
	ups int
	score int
}

func (bcd BasicCommentData) ToString() string {
	var buffer bytes.Buffer

	buffer.WriteString("{\n")
	buffer.WriteString("Author: " + bcd.author + "\n")
	buffer.WriteString("Subreddit: " + bcd.subreddit + "\n")
	buffer.WriteString("Body: " + bcd.body + "\n")
	buffer.WriteString("Score: " + strconv.Itoa(bcd.score) + "\n")
	buffer.WriteString("Ups: " + strconv.Itoa(bcd.ups) + "\n")
	buffer.WriteString("}")
	return buffer.String()
}

func getCommentDataFromLine(line []byte) BasicCommentData {
	var result BasicCommentData
	result.author = fastjson.GetString(line, "author")
	result.body = fastjson.GetString(line, "body")
	result.subreddit = fastjson.GetString(line, "subreddit")
	result.score = fastjson.GetInt(line, "score")
	result.ups = fastjson.GetInt(line, "ups")
	return result
}

func recurseBuildCompleteLine(reader *bufio.Reader) []byte {
	line, isPrefix, err := reader.ReadLine()
	if err != nil {
		log.Fatal(err)
	}

	if isPrefix {
		return append(line, recurseBuildCompleteLine(reader)...)
	} else {
		return line
	}
}

func main() {
	file, fileOpenErr := os.Open("D:/Torrents/2016/RC_2016-01/RC_2016-01")
	if fileOpenErr != nil {
		fmt.Print(fileOpenErr)
		os.Exit(0)
	}
	reader := bufio.NewReaderSize(file, 4096)

	fmt.Println("Opened file")

	var linesRead uint64 = 0

	startTime := time.Now()

	for {
		line := recurseBuildCompleteLine(reader)
		parsed := getCommentDataFromLine(line)

		if parsed.score > 100000 {
			fmt.Println("Parsed:\n" + parsed.ToString())
		}

		if linesRead % 100000 == 0 {
			dif := time.Now().Sub(startTime).String()
			fmt.Println("Total lines read: " + strconv.FormatUint(linesRead, 10) + " in " + dif)
		}
		linesRead++
	}
}
