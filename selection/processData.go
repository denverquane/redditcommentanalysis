package selection

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/denverquane/redditcommentanalysis/filesystem"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

const HundredThousand = 100000
const OneMillion = 1000000
const TenMillion = 10000000

type ProcessedSubredditStats struct {
	TotalComments int64
	Sentiment     BoxPlotStatistics
	WordLength    BoxPlotStatistics
	Karma         BoxPlotStatistics
}

func OpenExtractedSubredditDatafile(basedir, month, year, subreddit, extractedType string, progress *float64) ProcessedSubredditStats {
	retSummary := ProcessedSubredditStats{}

	str := basedir + "/Extracted/" + year + "/" + month + "/subreddit_" + subreddit + "_" + extractedType
	fmt.Println("Opening " + str)
	*progress = 0
	extractedDataFile, fileOpenErr := os.Open(str)
	if fileOpenErr != nil {
		log.Fatal("failed to open " + str)
	}
	totalLines, err := LineCounter(extractedDataFile)
	retSummary.TotalComments = int64(totalLines)
	if err != nil {
		log.Println(err)
	}
	if totalLines == 0 {
		log.Println("File has 0 comments; can't process!")
		return retSummary
	}
	sentiments := make([]float64, totalLines)
	wordLength := make([]float64, totalLines)
	karmas := make([]float64, totalLines)

	extractedDataFile.Seek(0, 0)
	extractedDataFileReader := bufio.NewReaderSize(extractedDataFile, 4096)

	lines := 0
	for {
		var tempComment map[string]string
		line := recurseBuildCompleteLine(extractedDataFileReader)
		if line == nil {
			break
		} else {
			err := json.Unmarshal(line, &tempComment)
			if err != nil {
				log.Fatal(err)
			}
			sentiments[lines], err = strconv.ParseFloat(tempComment["sentiment"], 64)
			wordLength[lines], err = strconv.ParseFloat(tempComment["wordlength"], 64)
			karmas[lines], err = strconv.ParseFloat(tempComment["score"], 64)
		}
		lines++
		*progress = 100.0 * (float64(lines) / float64(totalLines))
	}

	fmt.Println(strconv.Itoa(totalLines) + " total comments")

	retSummary.TotalComments = int64(totalLines)
	retSummary.WordLength = GetBoxPlotStats(wordLength)
	retSummary.Sentiment = GetBoxPlotStats(sentiments)
	retSummary.Karma = GetBoxPlotStats(karmas)

	DumpProcessedToCSV(basedir, month, year, subreddit, retSummary)
	MarshalToOutputFile(basedir, month, year, subreddit, retSummary)
	return retSummary
}

func DumpProcessedToCSV(basedir, month, year, subreddit string, processedStats ProcessedSubredditStats) {
	//filesystem.CreateSubdirectoryStructure("Processed", basedir, month, year)
	if !filesystem.DoesFolderExist(basedir + "/Processed") {
		filesystem.CreateFolder(basedir + "/Processed")
	}
	if !filesystem.DoesFolderExist(basedir + "/Processed/" + year) {
		filesystem.CreateFolder(basedir + "/Processed/" + year)
	}

	str := basedir + "/Processed/" + year + "/" + year + "_Summary.csv"

	var file *os.File
	var err error

	if filesystem.DoesFileExist(str) {
		file, err = os.OpenFile(str, os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			log.Println(err)
		}
		fileReader := bufio.NewReaderSize(file, 4096)

		for {
			line := recurseBuildCompleteLine(fileReader)
			if line == nil {
				break
			} else {
				entries := strings.Split(string(line), ",")
				if entries[0] == subreddit && entries[1] == month {
					log.Println("Processed sub data already in file, exiting!")
					file.Close()
					return
				}
			}
		}
	} else {
		file, err = os.Create(str)
		if err != nil {
			log.Println(err)
		}
		log.Println("Making new file: " + str)

		var header bytes.Buffer
		header.WriteString("cD#subreddit,")
		header.WriteString("iS#month,")
		header.WriteString("mC#avgwordLength,")
		header.WriteString("iC#medwordLength,")
		header.WriteString("mC#avgkarma,")
		header.WriteString("iC#medkarma,")
		header.WriteString("mC#avgsentiment,")
		header.WriteString("iC#medsentiment\n")
		file.Write(header.Bytes())
	}
	var buffer bytes.Buffer
	buffer.WriteString(subreddit + "," + month + ",")

	//for word := range processedStats.KeywordCommentKarmas {
	//	buffer.WriteString(word + ",")
	//}
	buffer.WriteString(strconv.FormatFloat(processedStats.WordLength.Average, 'f', 10, 64) + ",")
	buffer.WriteString(strconv.FormatFloat(processedStats.WordLength.Median, 'f', 10, 64) + ",")
	buffer.WriteString(strconv.FormatFloat(processedStats.Karma.Average, 'f', 10, 64) + ",")
	buffer.WriteString(strconv.FormatFloat(processedStats.Karma.Median, 'f', 10, 64) + ",")
	buffer.WriteString(strconv.FormatFloat(processedStats.Sentiment.Average, 'f', 10, 64) + ",")
	buffer.WriteString(strconv.FormatFloat(processedStats.Sentiment.Median, 'f', 10, 64) + ",")
	buffer.WriteString("\n")

	file.Write(buffer.Bytes())
	file.Close()
}

func MarshalToOutputFile(basedir, month, year, subreddit string, processedStats ProcessedSubredditStats) {
	filesystem.CreateSubdirectoryStructure("Processed", basedir, month, year)

	str := basedir + "/Processed/" + year + "/" + month + "/subreddit_" + subreddit

	if !filesystem.DoesFileExist(str) {
		file, err := os.Create(str)
		if err != nil {
			panic(err)
		}
		bytess, err2 := json.Marshal(processedStats)
		if err2 != nil {
			panic(err2)
		}
		file.Write(bytess)
		file.Close()
		log.Println("Successfully wrote " + subreddit + " to " + str)
	} else {
		log.Println("Subreddit " + subreddit + " already has a processed data file at " + str)
	}
}

func ScanDirForExtractedSubData(directory string, schema string) []string {
	subs := make([]string, 0)
	dirList, _ := ioutil.ReadDir(directory)
	for _, v := range dirList {
		if strings.Contains(v.Name(), "subreddit_") && strings.Contains(v.Name(), "_"+schema) && !strings.Contains(v.Name(), "_count") && !strings.Contains(v.Name(), ".tmp") {
			str := strings.Replace(v.Name(), "subreddit_", "", -1)
			str = strings.Replace(str, "_"+schema, "", -1)
			subs = append(subs, str)
		}
	}
	return subs
}

func ScanDirForProcessedSubData(directory string) []string {
	subs := make([]string, 0)
	dirList, _ := ioutil.ReadDir(directory)
	for _, v := range dirList {
		if strings.Contains(v.Name(), "subreddit_") && !strings.Contains(v.Name(), ".tmp") && !strings.Contains(v.Name(), ".csv") {
			str := strings.Replace(v.Name(), "subreddit_", "", -1)
			subs = append(subs, str)
		}
	}
	return subs
}
