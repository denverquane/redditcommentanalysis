package selection

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/denverquane/redditcommentanalysis/filesystem"
	"github.com/valyala/fastjson"
	"log"
	"os"
	"regexp"
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
	filter, _ := regexp.Compile("[^a-zA-Z0-9 ']+")

	for key, v := range keyTypes {
		if v == "str" {
			if key == "body" {
				result[key] = filter.ReplaceAllString(filterStopWordsFromString(strings.ToLower(fastjson.GetString(line, key))), "")
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
	var extractedCommentCount int64 = 0
	var commentsInRawFile int64
	criteria = strings.ToLower(criteria)
	value = strings.ToLower(value)

	filesystem.CreateSubdirectoryStructure(basedir, month, year)

	outputFilePath := basedir + "/Extracted/" + year + "/" + month + "/" + criteria + "_" + value + "_" + schema.name
	rawDataFilePath := basedir + "/RC_" + year + monthToIntString(month)

	commentsInRawFile = readInCommentCountMetadata(rawDataFilePath)

	if _, err := os.Stat(outputFilePath); !os.IsNotExist(err) {
		fmt.Println("Found cached data for " + outputFilePath)
		return commentsInRawFile
	} else {
		fmt.Println("No cached data found for " + outputFilePath)
	}

	outputFileWriter, e := os.Create(outputFilePath + ".tmp")
	if e != nil {
		log.Println(e)
	}
	//Note: Not deferring a outputfile close, because we need to close/rename when extraction is done

	rawDataFile, fileOpenErr := os.Open(rawDataFilePath)
	if fileOpenErr != nil {
		fmt.Print(fileOpenErr)
		os.Exit(0)
	} else {
		fmt.Println("Opened file " + rawDataFilePath)
	}

	rawDataFileReader := bufio.NewReaderSize(rawDataFile, 4096)

	var linesRead int64 = 0
	startTime := time.Now()
	tempTime := startTime
	for {
		line := recurseBuildCompleteLine(rawDataFileReader)
		if line == nil {
			fmt.Println("Lines: " + strconv.FormatInt(linesRead, 10))
			log.Println("Encountered error; concluding analysis")
			break
		}
		linesRead++

		if strings.Contains(strings.ToLower(string(line)), "\""+criteria+"\":\""+value+"\"") {
			parsed := getCommentDataFromLine(line, schema.schema)
			marshalled, e := json.Marshal(parsed)
			if e != nil {
				log.Println(e)
			} else {
				outputFileWriter.Write(marshalled)
				outputFileWriter.Write([]byte("\n"))
				extractedCommentCount++
			}
		}

		if linesRead%HundredThousand == 0 {
			progressStr := ""
			if commentsInRawFile != 0 {
				*progress = (float64(linesRead) / float64(commentsInRawFile)) * 100.0
				progressStr = strconv.FormatFloat(*progress, 'f', 2, 64) + "%"
			} else {
				progressStr = strconv.FormatInt(linesRead, 10) + " lines total"
			}
			fmt.Println("Processed 100k lines in " + time.Now().Sub(tempTime).String() + " (" + progressStr + ")")
			tempTime = time.Now()
		}
	}
	dif := time.Now().Sub(startTime).String()
	tempStr := "Took " + dif + " to search " + strconv.FormatInt(linesRead, 10) + " comments of file " + rawDataFilePath + "\n"
	fmt.Println(tempStr)

	if extractedCommentCount == 0 {
		log.Println("Found 0 comments for " + criteria + ":" + value + " in " + month + ", exiting extraction!")
		return commentsInRawFile
	} else {
		log.Println("Extracted " + strconv.FormatInt(extractedCommentCount, 10) + " comments for " +
			criteria + " = " + value + " in " + month + "/" + year)
	}
	*progress = 100.0
	if commentsInRawFile == 0 {
		log.Println("Never read the linecount from a file; writing to file now")
		f, err := os.Create(rawDataFilePath + "_count")
		if err != nil {
			log.Println(err)
		}
		f.Write([]byte(strconv.FormatInt(int64(linesRead), 10)))
		f.Close()
	}

	outputFileWriter.Close()
	err := os.Rename(outputFilePath+".tmp", outputFilePath)
	log.Println(err)

	return linesRead
}
