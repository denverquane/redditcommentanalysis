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
	"sort"
	"strconv"
	"strings"
)

const HundredThousand = 100000
const OneMillion = 1000000
const TenMillion = 10000000
const SampleComments = 40000

type ProcessedSubredditStats struct {
	TotalComments int64
	Polarity      BoxPlotStatistics
	Subjectivity  BoxPlotStatistics
	WordLength    BoxPlotStatistics
	Karma         BoxPlotStatistics
	Diversity     float64
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

	diversity := GetWordDiversity(SampleComments, totalLines, extractedDataFile)
	//fmt.Println("Word diversity for " + subreddit + " is " + strconv.FormatFloat(diversity, 'f', 10, 64))
	polarity := make([]float64, totalLines)
	subjectivity := make([]float64, totalLines)
	wordLength := make([]float64, totalLines)
	karmas := make([]float64, totalLines)

	extractedDataFile.Seek(0, 0)
	extractedDataFileReader := bufio.NewReaderSize(extractedDataFile, 4096)

	lines := 0
	filesystem.CreateSubdirectoryStructure("Processed", basedir, month, year)

	// Opening a file is a huge bottleneck, so just create/open it once and hand it around
	classFileStr := basedir + "/Processed/" + year + "/" + month + "/classification.csv"
	var classFile *os.File

	if !filesystem.DoesFileExist(classFileStr) {
		classFile, err = os.Create(classFileStr)

		if err != nil {
			panic(err)
		}
		var header bytes.Buffer
		header.WriteString("cD#subreddit,")
		//header.WriteString("iS#month,")
		header.WriteString("mD#wordLength,")
		header.WriteString("mD#karma,")
		header.WriteString("mC#polarity,")
		header.WriteString("mC#subjectivity\n")
		classFile.Write(header.Bytes())

	} else {
		classFile, err = os.OpenFile(classFileStr, os.O_RDWR|os.O_APPEND, 0660)
	}
	defer classFile.Close()

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
			polarity[lines], err = strconv.ParseFloat(tempComment["polarity"], 64)
			subjectivity[lines], err = strconv.ParseFloat(tempComment["subjectivity"], 64)
			wordLength[lines], err = strconv.ParseFloat(tempComment["wordlength"], 64)
			karmas[lines], err = strconv.ParseFloat(tempComment["score"], 64)

			DumpLineToClassificationFile(subreddit, wordLength[lines], karmas[lines], polarity[lines], subjectivity[lines], classFile)
			//dump to file here
		}
		lines++
		*progress = 100.0 * (float64(lines) / float64(totalLines))
	}

	fmt.Println(strconv.Itoa(totalLines) + " total comments")

	retSummary.TotalComments = int64(totalLines)

	sort.Float64s(karmas)
	sort.Float64s(wordLength)
	sort.Float64s(polarity)
	sort.Float64s(subjectivity)

	retSummary.WordLength = GetBoxPlotStats(wordLength)
	retSummary.Polarity = GetBoxPlotStats(polarity)
	retSummary.Subjectivity = GetBoxPlotStats(subjectivity)
	retSummary.Karma = GetBoxPlotStats(karmas)

	retSummary.Diversity = diversity

	//DumpProcessedToCSV(basedir, month, year, subreddit, retSummary)
	MarshalToOutputFile(basedir, month, year, subreddit, retSummary)
	return retSummary
}

func GetWordDiversity(numComments, totalComments int, file *os.File) float64 {
	file.Seek(0, 0)
	fileReader := bufio.NewReaderSize(file, 4096)
	lines := 0
	divider := int(float64(totalComments) / float64(numComments))
	if divider == 0 {
		divider = 1
	}
	allWords := make(map[string]int64)
	totalWords := 0
	sampled := 0
	wordLengths := make([]float64, numComments)
	for {
		var tempComment map[string]string
		line := recurseBuildCompleteLine(fileReader)
		if line == nil || sampled == numComments {
			break
		} else if lines%divider == 0 {
			err := json.Unmarshal(line, &tempComment)
			if err != nil {
				log.Fatal(err)
			}

			body, _ := tempComment["body"]
			words := strings.Fields(body)
			wordLengths[sampled] = float64(len(words))
			for _, v := range words { //all the words of the comment
				totalWords++
				if count, ok := allWords[v]; ok { //if we've seen the word before
					allWords[v] = count + 1
				} else if !strings.Contains(v, "http") {
					allWords[v] = 1 //haven't seen this word before
				}
			}
			sampled++
		}
		lines++
	}
	uniqueWords := len(allWords)

	//fmt.Println(strconv.Itoa(uniqueWords) + " unique words, " + strconv.Itoa(totalWords) + " total words")
	//
	//fmt.Println("Sampled " + strconv.Itoa(sampled) + " comments")
	return float64(uniqueWords) / float64(sampled)
}

//TODO Broken!!!
func DumpProcessedToCSV(basedir, month, year, subreddit string, processedStats ProcessedSubredditStats) bool {
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
					return true
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
	//buffer.WriteString(strconv.FormatFloat(processedStats.Sentiment.Average, 'f', 10, 64) + ",")
	//buffer.WriteString(strconv.FormatFloat(processedStats.Sentiment.Median, 'f', 10, 64) + ",")
	buffer.WriteString("\n")

	file.Write(buffer.Bytes())
	file.Close()
	return false
}

func DumpLineToClassificationFile(sub string, wordLength, karma, polarity, subjectivity float64, file *os.File) {
	var buffer bytes.Buffer
	buffer.WriteString(sub + ",")
	buffer.WriteString(strconv.FormatFloat(wordLength, 'f', 0, 64) + ",")
	buffer.WriteString(strconv.FormatFloat(karma, 'f', 0, 64) + ",")
	buffer.WriteString(strconv.FormatFloat(polarity, 'f', 10, 64) + ",")
	buffer.WriteString(strconv.FormatFloat(subjectivity, 'f', 10, 64))
	buffer.WriteString("\n")
	file.Write(buffer.Bytes())
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
