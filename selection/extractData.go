package selection

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/denverquane/redditcommentanalysis/filesystem"
	"github.com/valyala/fastjson"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

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

func getCommentDataFromLine(line []byte, keyTypes map[string]string) map[string]string {
	result := make(map[string]string, 0)

	for key, v := range keyTypes {
		if v == "str" {
			if key == "body" {
				result[key] = filterStopWordsFromString(strings.ToLower(fastjson.GetString(line, key)))
			} else {
				result[key] = strings.ToLower(fastjson.GetString(line, key))
			}
		} else if v == "int" {
			result[key] = strconv.Itoa(fastjson.GetInt(line, key))
		} else {
			log.Fatal("Undetected type: " + v)
		}
	}
	return result
}

const percentPerMonth = (1.0 / 12.0) * 100.0

func ExtractCriteriaDataToFile(criteria string, value string, year string, month string, basedir string, schema commentSchema, progress *float64) int64 {
	relevantComments := make([]map[string]string, 0)
	var summary int64
	criteria = strings.ToLower(criteria)
	value = strings.ToLower(value)

	if !filesystem.DoesFolderExist(basedir + "/Extracted") {
		filesystem.CreateFolder(basedir + "/Extracted")
	}
	if !filesystem.DoesFolderExist(basedir + "/Extracted/" + year) {
		filesystem.CreateFolder(basedir + "/Extracted/" + year)
	}
	if !filesystem.DoesFolderExist(basedir + "/Extracted/" + year + "/" + month) {
		filesystem.CreateFolder(basedir + "/Extracted/" + year + "/" + month)
	}

	outputFilePath := basedir + "/Extracted/" + year + "/" + month + "/" + criteria + "_" + value + "_" + schema.name

	if _, err := os.Stat(outputFilePath); !os.IsNotExist(err) {
		linecountFilePath := outputFilePath + "_count"
		if _, err := os.Stat(linecountFilePath); !os.IsNotExist(err) {
			var count int64
			plan, fileOpenErr := ioutil.ReadFile(linecountFilePath)
			if fileOpenErr != nil {
				log.Fatal("failed to open " + linecountFilePath)
			}
			err := json.Unmarshal(plan, &count)
			if err != nil {
				log.Fatal(err)
			}
			summary = count
		}
		fmt.Println("Found cached data for " + outputFilePath)
		return summary
	} else {
		fmt.Println("No cached data found for " + outputFilePath)
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(basedir + "/RC_" + year + monthToIntString(month)))
	file, fileOpenErr := os.Open(buffer.String())
	if fileOpenErr != nil {
		fmt.Print(fileOpenErr)
		os.Exit(0)
	} else {
		fmt.Println("Opened file " + buffer.String())
	}

	var totalLines int64
	metafile, fileOpenErr2 := os.Open(buffer.String() + "_meta.txt")
	if fileOpenErr2 != nil {
		totalLines = 0
		metafile.Close()
	} else {
		rdr := bufio.NewReader(metafile)
		line, _, _ := rdr.ReadLine()
		totalLines, _ = strconv.ParseInt(string(line), 10, 64)
	}

	reader := bufio.NewReaderSize(file, 4096)

	var linesRead int64 = 0
	startTime := time.Now()
	tempTime := startTime
	for {
		line := recurseBuildCompleteLine(reader)
		if line == nil {
			fmt.Println("Lines: " + strconv.FormatInt(linesRead, 10))
			log.Println("Encountered error; concluding analysis")
			break
		}
		linesRead++

		if strings.Contains(strings.ToLower(string(line)), "\""+criteria+"\":\""+value+"\"") {
			parsed := getCommentDataFromLine(line, schema.schema)
			relevantComments = append(relevantComments, parsed)
		}

		if linesRead%HundredThousand == 0 {
			if totalLines != 0 {
				*progress = (float64(linesRead) / float64(totalLines)) * 100.0
				fmt.Println("Percent complete with " + month + ": " + strconv.FormatFloat(*progress, 'f', 2, 64) + "%")
			} else {
				fmt.Println("Lines complete: " + strconv.FormatInt(linesRead, 10))
			}
			fmt.Println("Processed 100k lines in " + time.Now().Sub(tempTime).String())
			tempTime = time.Now()
		}
	}
	dif := time.Now().Sub(startTime).String()
	tempStr := "Took " + dif + " to search " + strconv.FormatInt(linesRead, 10) + " comments of file " + buffer.String() + "\n"
	if len(relevantComments) == 0 {
		log.Println("Found 0 comments for " + criteria + ":" + value + " in " + month + ", exiting extraction!")
		return summary
	}
	summary = linesRead
	fmt.Println(tempStr)
	dumpDataToFilepath(relevantComments, outputFilePath)
	*progress = 100.0

	return summary
}
