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
	KeywordCommentTallies map[string]float64 //How many comments contain the keyword
	KeywordCommentKarmas  map[string]float64 //total karma for unique occurrences of the keyword
	AverageSentiment      float64
}

func OpenExtractedSubredditDatafile(basedir, month, year, subreddit, extractedType string, progress *float64) ProcessedSubredditStats {
	retSummary := ProcessedSubredditStats{make(map[string]float64, 0), make(map[string]float64), 0}
	var commentData []map[string]string
	str := basedir + "/Extracted/" + year + "/" + month + "/subreddit_" + subreddit + "_" + extractedType
	fmt.Println("Opening " + str)
	*progress = 0
	extractedDataFile, fileOpenErr := os.Open(str)
	if fileOpenErr != nil {
		log.Fatal("failed to open " + str)
	}
	totalLines, err := LineCounter(extractedDataFile)
	if err != nil {
		log.Println(err)
	}
	if totalLines == 0 {
		log.Println("File has 0 comments; can't process!")
		return retSummary
	}
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
			commentData = append(commentData, tempComment)
		}
		lines++
		*progress = 50.0 * (float64(lines) / float64(totalLines))
	}

	fmt.Println(strconv.Itoa(len(commentData)) + " total comments")
	*progress = 60

	fmt.Println("Tallying word and karma counts...")
	sortedTallies, sortedKarmas := tallyWordOccurrencesAndSort(commentData)
	for _, v := range sortedTallies[:TotalTallyAndKarmaRecords] {
		percent := (float64(v.TotalCount) / float64(len(commentData))) * 100.0
		str := v.Word + ": " + strconv.FormatInt(v.TotalCount, 10) + " comment occurrences (" +
			strconv.FormatFloat(percent, 'f', 3, 64) + "%)"
		retSummary.KeywordCommentTallies[v.Word] = percent
		fmt.Println(str)
	}
	*progress = 75
	fmt.Println("getting sortedKarmas")

	for _, v := range sortedTallies[:TotalTallyAndKarmaRecords] {
		for _, karma := range sortedKarmas {
			if karma.Word == v.Word {
				str := karma.Word + ": " + strconv.FormatInt(karma.TotalKarma, 10) + " karma total"
				retSummary.KeywordCommentKarmas[karma.Word] = float64(karma.TotalKarma) / float64(karma.TotalCount)
				fmt.Println(str)
			}
		}
	}
	*progress = 90
	fmt.Println("getting sentiments")
	var sentTotal = 0.0
	var totalRead = 0
	for _, v := range commentData {
		str := v["sentiment"]
		if str == "" {
			continue
		}
		f64, err := strconv.ParseFloat(str, 64)
		if err != nil {
			continue
		}
		if f64 != 0 {
			sentTotal += f64
			totalRead++
		}

	}
	avgSent := sentTotal / float64(totalRead)
	fmt.Println("Subreddit " + subreddit + " avg sentiment: " + strconv.FormatFloat(avgSent, 'f', 10, 64))
	retSummary.AverageSentiment = avgSent

	*progress = 100
	DumpProcessedToCSV(basedir, month, year, subreddit, extractedType, retSummary)
	return retSummary
}

func DumpProcessedToCSV(basedir, month, year, subreddit, extractedType string, processedStats ProcessedSubredditStats) {
	//filesystem.CreateSubdirectoryStructure("Processed", basedir, month, year)
	if !filesystem.DoesFolderExist(basedir + "/Processed") {
		filesystem.CreateFolder(basedir + "/Processed")
	}
	if !filesystem.DoesFolderExist(basedir + "/Processed/" + year) {
		filesystem.CreateFolder(basedir + "/Processed/" + year)
	}
	str := basedir + "/Processed/" + year + "/subreddit_" + subreddit + "_" + extractedType + ".csv"

	var file *os.File
	var err error

	if filesystem.DoesFileExist(str) {
		file, err = os.OpenFile(str, os.O_APPEND|os.O_WRONLY, 0600)
		log.Println("Processed csv file already exists; appending to " + str)
	} else {
		file, err = os.Create(str)
		log.Println("Making new file: " + str)
	}

	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	var buffer bytes.Buffer
	buffer.WriteString(subreddit + "," + month + ",")

	buffer.WriteString("\"")
	for word := range processedStats.KeywordCommentKarmas {
		buffer.WriteString(word + ",")
	}
	buffer.WriteString("\",")
	buffer.WriteString(strconv.FormatFloat(processedStats.AverageSentiment, 'f', 10, 64) + ",")
	buffer.WriteString("\n")

	file.Write(buffer.Bytes())
}

func CombineAllToSingleCSV(basedir, year, extractedType string) {
	str := basedir + "/Processed/" + year + "/"

	f, err := os.Open(str)
	if err != nil {
		log.Fatal(err)
	}

	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	output, err2 := os.Create(str + "Summary.csv")
	if err2 != nil {
		log.Fatal(err2)
	}
	defer output.Close()

	for _, v := range files {
		if !v.IsDir() && strings.HasSuffix(v.Name(), ".csv") && strings.HasPrefix(v.Name(), "subreddit_") {
			csv, err := os.Open(str + v.Name())
			if err != nil {
				log.Println(err)
			}
			rawDataFileReader := bufio.NewReaderSize(csv, 4096)

			for {
				line := recurseBuildCompleteLine(rawDataFileReader)
				if line == nil {
					break
				} else {
					output.WriteString(string(line) + "\n")
				}
			}
			csv.Close()
		}
	}
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

func average(ints []int64) float64 {
	var sum int64 = 0
	for _, v := range ints {
		sum += v
	}
	return float64(sum) / float64(len(ints))
}

func faverage(floats []float64) float64 {
	var sum float64 = 0
	for _, v := range floats {
		sum += v
	}
	return sum / float64(len(floats))
}
