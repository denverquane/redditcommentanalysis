package selection

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/valyala/fastjson"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const HundredThousand = 100000
const OneMillion = 1000000

type SearchParams struct {
	year                string
	month               string
	lineLimit           uint64
	formattedORQueries  []string
	formattedANDQueries []string
}

func MakeSimpleSearchParams(year string, month string) SearchParams {
	return SearchParams{year, month, OneMillion,
		make([]string, 0), make([]string, 0)}
}

func (params *SearchParams) AddANDCriteria(key string, val string) *SearchParams {
	params.formattedANDQueries = append(params.formattedANDQueries, makeFormattedSearchPair(key, val))
	return params
}

func (params *SearchParams) AddORCriteria(key string, val string) *SearchParams {
	params.formattedORQueries = append(params.formattedORQueries, makeFormattedSearchPair(key, val))
	return params
}

func makeFormattedSearchPair(key string, val string) string {
	return "\"" + key + "\":\"" + val + "\""
}

type BasicCommentData struct {
	author    string
	subreddit string
	body      string
	score     int
}

func (bcd BasicCommentData) ToString() string {
	var buffer bytes.Buffer
	buffer.WriteString("{\n")
	buffer.WriteString("Author: " + bcd.author + "\n")
	buffer.WriteString("Subreddit: " + bcd.subreddit + "\n")
	buffer.WriteString("Body: " + bcd.body + "\n")
	buffer.WriteString("Score: " + strconv.Itoa(bcd.score) + "\n")
	buffer.WriteString("}")
	return buffer.String()
}

func getCommentDataFromLine(line []byte) BasicCommentData {
	var result BasicCommentData
	result.author = fastjson.GetString(line, "author")
	result.body = fastjson.GetString(line, "body")
	result.subreddit = fastjson.GetString(line, "subreddit")
	result.score = fastjson.GetInt(line, "score")
	return result
}

//TODO explore speedups related to unravelling the recursion
func recurseBuildCompleteLine(reader *bufio.Reader) []byte {
	line, isPrefix, err := reader.ReadLine()
	if err != nil {
		log.Println(err)
		return nil
	}

	if isPrefix {
		return append(line, recurseBuildCompleteLine(reader)...)
	} else {
		return line
	}
}

func monthToIntString(month string) string {
	switch month {
	case "Jan":
		return "-01"
	case "Feb":
		return "-02"
	case "Mar":
		return "-03"
	case "Apr":
		return "-04"
	case "May":
		return "-05"
	case "Jun":
		return "-06"
	case "Jul":
		return "-07"
	case "Aug":
		return "-08"
	case "Sep":
		return "-09"
	case "Oct":
		return "-10"
	case "Nov":
		return "-11"
	case "Dec":
		return "-12"
	default:
		log.Println("Invalid month supplied; defaulting to january")
		return "-01"
	}
}

func OpenAndSearchFile(params SearchParams, baseDataDirectory string) []BasicCommentData {
	var buffer bytes.Buffer
	buffer.Write([]byte(baseDataDirectory))
	buffer.Write([]byte("/RC_" + params.year + monthToIntString(params.month)))
	file, fileOpenErr := os.Open(buffer.String())

	if fileOpenErr != nil {
		fmt.Print(fileOpenErr)
		os.Exit(0)
	}
	reader := bufio.NewReaderSize(file, 4096)

	fmt.Println("Opened file " + buffer.String())

	var linesRead uint64 = 0
	relevantComments := make([]BasicCommentData, 0)

	startTime := time.Now()

	for {
		if linesRead > 0 && linesRead == params.lineLimit {
			log.Println("Reached line limit, concluding analysis")
			break
		}

		line := recurseBuildCompleteLine(reader)
		if line == nil {
			log.Println("Encountered error; concluding analysis")
			break
		}
		linesRead++

		if meetsCriteria(string(line), params.formattedORQueries, params.formattedANDQueries) {
			parsed := getCommentDataFromLine(line)
			relevantComments = append(relevantComments, parsed)
		}

		if linesRead%HundredThousand == 0 {
			fmt.Println("Read " + strconv.FormatUint(linesRead, 10) + " lines so far")
		}
	}
	dif := time.Now().Sub(startTime).String()
	fmt.Println("Took " + dif + " to search " + strconv.FormatUint(linesRead, 10) + " comments of file " + buffer.String())
	return relevantComments
}

func meetsCriteria(line string, ORs []string, ANDs []string) bool {
	var metCriteria bool
	if len(ORs) > 0 {
		metCriteria = false
	} else {
		metCriteria = true //no ORs to test- automatically valid until AND step
	}

	for _, v := range ORs {
		if strings.Contains(line, v) {
			metCriteria = true //met a SINGLE one of the OR criteria; continues on
			break
		}
	}
	if !metCriteria { //failed all the ORs
		return false
	}

	for _, v := range ANDs {
		if !strings.Contains(line, v) {
			return false
		}
	}
	return true
}
