package selection

import (
	"bufio"
	"encoding/json"
	"fmt"
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
}

func OpenExtractedSubredditDatafile(basedir, month, year, subreddit, extractedType string, progress *float64) ProcessedSubredditStats {
	retSummary := ProcessedSubredditStats{make(map[string]float64, 0), make(map[string]float64)}
	var commentData []map[string]string
	str := basedir + "/Extracted/" + year + "/" + month + "/subreddit_" + subreddit + "_" + extractedType
	fmt.Println("Opening " + str)

	extractedDataFile, fileOpenErr := os.Open(str)
	if fileOpenErr != nil {
		log.Fatal("failed to open " + str)
	}
	totalLines, err := LineCounter(extractedDataFile)
	if err != nil {
		log.Println(err)
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
	tallies, karmas := tallyWordOccurrences(commentData)
	for _, v := range tallies[:TotalTallyAndKarmaRecords] {
		percent := (float64(v.TotalCount) / float64(len(commentData))) * 100.0
		str := v.Word + ": " + strconv.FormatInt(v.TotalCount, 10) + " comment occurrences (" +
			strconv.FormatFloat(percent, 'f', 3, 64) + "%)"
		retSummary.KeywordCommentTallies[v.Word] = percent
		fmt.Println(str)
	}
	*progress = 75
	fmt.Println("getting karmas")

	for _, v := range tallies[:TotalTallyAndKarmaRecords] {
		for _, karma := range karmas {
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

	*progress = 100
	return retSummary
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

func tallyWordOccurrences(comments []map[string]string) (TallyList, KarmaList) {
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
