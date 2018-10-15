package selection

import (
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

func OpenExtractedSubredditDatafile(basedir string, subreddit string, extractedType string, progress *float64) string {
	retStr := ""
	var commentData []map[string]string

	for i, v := range AllMonths {
		fmt.Println("Reading " + v)
		var tempCommData []map[string]string
		str := basedir + "/" + v + "/" + "subreddit_" + subreddit + "_" + extractedType
		plan, fileOpenErr := ioutil.ReadFile(str)
		if fileOpenErr != nil {
			log.Fatal("failed to open " + str)
		}
		err := json.Unmarshal(plan, &tempCommData)
		if err != nil {
			log.Println(err)
		}
		fmt.Println(subreddit + " has " + strconv.FormatInt(int64(len(tempCommData)), 10) + " comments in " + v)
		commentData = append(commentData, tempCommData...)
		*progress = 0.5 * ((float64(i) + 1.0) * percentPerMonth)
	}
	fmt.Println(strconv.Itoa(len(commentData)) + " total comments")
	*progress = 50

	fmt.Println("Tallying word and karma counts...")
	tallies, karmas := tallyWordOccurrences(commentData)
	for _, v := range tallies[:10] {
		percent := (float64(v.TotalCount) / float64(len(commentData))) * 100.0
		str := v.Word + ": " + strconv.FormatInt(v.TotalCount, 10) + " comment occurrences (" +
			strconv.FormatFloat(percent, 'f', 3, 64) + "%)"
		retStr += str + "\n"
		fmt.Println(str)
	}
	*progress = 75
	fmt.Println("getting karmas")
	for _, karma := range karmas[:10] {
		str := karma.Word + ": " + strconv.FormatInt(karma.TotalKarma, 10) + " karma total"
		retStr += str + "\n"
		fmt.Println(str)
	}
	*progress = 100
	return retStr
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

const MaxGoRoutines = 8

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

func dumpDataToFilepath(data []map[string]string, filePath string) {
	f, err := os.Create(filePath)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	f2, err2 := os.Create(filePath + "_count")
	if err2 != nil {
		log.Println(err2)
	}
	defer f2.Close()
	length := strconv.FormatInt(int64(len(data)), 10)
	f2.Write([]byte(length))

	bytess, err3 := json.Marshal(data)
	if err3 != nil {
		log.Println(err3)
	}
	f.Write(bytess)
}

func ScanDirForExtractedSubData(directory string, schema string) []string {
	subs := make([]string, 0)
	dirList, _ := ioutil.ReadDir(directory)
	for _, v := range dirList {
		if strings.Contains(v.Name(), "subreddit_") && strings.Contains(v.Name(), "_"+schema) && !strings.Contains(v.Name(), "_count") {
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
