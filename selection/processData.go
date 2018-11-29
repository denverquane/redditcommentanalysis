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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const HundredThousand = 100000
const OneMillion = 1000000
const TenMillion = 10000000

const TotalTallyAndKarmaRecords = 50

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
		*progress = 50.0 * (float64(lines) / float64(totalLines))
	}

	fmt.Println(strconv.Itoa(totalLines) + " total comments")

	*progress = 90
	fmt.Println("getting sentiments")

	retSummary.TotalComments = int64(totalLines)
	retSummary.WordLength = GetBoxPlotStats(wordLength)
	retSummary.Sentiment = GetBoxPlotStats(sentiments)
	retSummary.Karma = GetBoxPlotStats(karmas)

	*progress = 100
	DumpProcessedToCSV(basedir, month, year, subreddit, retSummary)
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
		header.WriteString("mC#wordLength,")
		header.WriteString("mC#karma,")
		header.WriteString("mC#sentiment\n")
		file.Write(header.Bytes())
	}
	var buffer bytes.Buffer
	buffer.WriteString(subreddit + "," + month + ",")

	//for word := range processedStats.KeywordCommentKarmas {
	//	buffer.WriteString(word + ",")
	//}
	buffer.WriteString(strconv.FormatFloat(processedStats.WordLength.Median, 'f', 10, 64) + ",")
	buffer.WriteString(strconv.FormatFloat(processedStats.Karma.Median, 'f', 10, 64) + ",")
	buffer.WriteString(strconv.FormatFloat(processedStats.Sentiment.Median, 'f', 10, 64) + ",")
	buffer.WriteString("\n")

	file.Write(buffer.Bytes())
	file.Close()
}

func sortKarma(karmaCounts map[string]IntPair) KarmaList {
	pl := make(KarmaList, len(karmaCounts))
	i := 0
	for k, v := range karmaCounts {
		//karmaPerOccurrence := float64(v.karma) / float64(v.tally)
		pl[i] = KeywordStats{k, v.karma, v.tally, []string{""}}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

func sortTallies(tallies map[string]IntPair) TallyList {
	pl := make(TallyList, len(tallies))
	i := 0
	for k, v := range tallies {
		pl[i] = KeywordStats{k, v.karma, v.tally, []string{""}}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type KeywordStats struct {
	Word       string
	TotalKarma int64
	TotalCount int64
	CommentIDs []string //TODO Might be helpful for retrieving a specific comment that needs manual viewing
}

type IntPair struct {
	tally int64
	karma int64
}

type KarmaList []KeywordStats

func (p KarmaList) Len() int           { return len(p) }
func (p KarmaList) Less(i, j int) bool { return p[i].TotalKarma < p[j].TotalKarma }
func (p KarmaList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type TallyList []KeywordStats

func (p TallyList) Len() int           { return len(p) }
func (p TallyList) Less(i, j int) bool { return p[i].TotalCount < p[j].TotalCount }
func (p TallyList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

const MaxGoRoutines = 4

func tallyWordOccurrencesAndSort(comments []map[string]string) (TallyList, KarmaList) {
	tallies := make(map[string]IntPair)
	var mux sync.Mutex

	fmt.Println("Using " + strconv.Itoa(MaxGoRoutines) + " workers")

	perWorker := len(comments) / MaxGoRoutines

	c := make(chan bool, MaxGoRoutines)

	start := time.Now()
	for i := 0; i < MaxGoRoutines-1; i++ {
		fmt.Println("from " + strconv.Itoa(i*perWorker) + " to " + strconv.Itoa((i+1)*perWorker))
		go tallyCommentsWorker(comments[i*perWorker:(i+1)*perWorker], &mux, &tallies, c)
	}
	fmt.Println("from " + strconv.Itoa((MaxGoRoutines-1)*perWorker) + " to " + strconv.Itoa(len(comments)))
	go tallyCommentsWorker(comments[(MaxGoRoutines-1)*perWorker:], &mux, &tallies, c)

	for i := 0; i < MaxGoRoutines; i++ {
		<-c
	}

	total := time.Now().Sub(start)
	fmt.Println("Total time: " + total.String())

	return sortTallies(tallies), sortKarma(tallies)
}

func tallyCommentsWorker(comments []map[string]string, lock *sync.Mutex, tallies *map[string]IntPair, c chan bool) {
	fmt.Println("Started worker!")
	filter, _ := regexp.Compile("[^a-zA-Z0-9]+")
	for _, comment := range comments {
		words := strings.Fields(comment["body"])
		wordPresence := make(map[string]bool, len(words))
		for _, word := range words {
			word = strings.ToLower(word)
			word = filter.ReplaceAllString(word, "")
			if IsStopWord(word) || word == "" {
				continue
			}
			// wordPresence ensures we don't count repeat occurrences of a word within a single comment
			if _, ok := wordPresence[word]; !ok {
				wordPresence[word] = true
				karma, err := strconv.ParseInt(comment["score"], 10, 64)
				if err != nil {
					log.Fatal(err)
				}
				lock.Lock()
				if _, ok := (*tallies)[word]; ok {
					pair := (*tallies)[word]
					pair.karma += int64(karma)
					pair.tally++
					(*tallies)[word] = pair
				} else {
					(*tallies)[word] = IntPair{tally: 1, karma: int64(karma)}
				}
				lock.Unlock()
			}
		}
	}
	c <- true
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
