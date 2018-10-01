package selection

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/valyala/fastjson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const HundredThousand = 100000
const OneMillion = 1000000
const TenMillion = 10000000

type SearchParams struct {
	year             string
	months           []string
	lineLimit        uint64
	stringORQueries  []string
	stringANDQueries []string
	intORQueries     []IntCriteria
	intANDQueries    []IntCriteria
}

func (params SearchParams) ToString() string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	buffer.WriteString("\nyear: " + params.year)
	buffer.WriteString("\nmonths: \n[")
	for _, v := range params.months {
		buffer.WriteString("\n" + v)
	}
	buffer.WriteString("\n]")
	buffer.WriteString("\nlinelimit: " + strconv.FormatUint(params.lineLimit, 10))
	buffer.WriteString("\nORQueries: \n[")
	for _, v := range params.stringORQueries {
		buffer.WriteString("\n" + v)
	}
	buffer.WriteString("\n]")
	buffer.WriteString("\nANDQueries: \n[")
	for _, v := range params.stringANDQueries {
		buffer.WriteString("\n" + v)
	}
	buffer.WriteString("\n]")
	return buffer.String()
}

func (params SearchParams) ValuesToString() string {
	var buffer bytes.Buffer
	buffer.WriteString(params.year + "_")
	for _, v := range params.months {
		buffer.WriteString(v + "_")
	}
	buffer.WriteString(strconv.FormatUint(params.lineLimit, 10) + "_")
	for _, v := range params.stringORQueries {
		str := strings.Replace(v, "\"", "", -1) + "_"
		buffer.WriteString(strings.Replace(str, ":", "", -1) + "_")
	}

	for _, v := range params.stringANDQueries {
		str := strings.Replace(v, "\"", "", -1) + "_"
		buffer.WriteString(strings.Replace(str, ":", "", -1) + "_")
	}
	return buffer.String()
}

type SimpleSearchCriteria struct {
	Year             string
	Months           []string
	LineLimit        int
	StringORQueries  []string
	StringANDQueries []string
}

func InterpretFromHttp(r *http.Request) *SearchParams {
	params := MakeEmptySearchParams()
	var message SimpleSearchCriteria
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&message); err != nil {
		fmt.Println(err)
	}
	fmt.Println(message)
	params.year = message.Year
	params.months = message.Months
	params.lineLimit = uint64(message.LineLimit)
	params.stringORQueries = message.StringORQueries
	params.stringANDQueries = message.StringANDQueries
	fmt.Println(params)
	return params
}

func MakeEmptySearchParams() *SearchParams {
	return &SearchParams{"", make([]string, 0), 0,
		make([]string, 0), make([]string, 0),
		make([]IntCriteria, 0), make([]IntCriteria, 0)}
}

func MakeSimpleSearchParams(year string, months []string, limit uint64, ORs []string, ANDs []string) *SearchParams {
	return &SearchParams{year, months, limit,
		ORs, ANDs,
		make([]IntCriteria, 0), make([]IntCriteria, 0)}
}

func (param SearchParams) IsReady() (bool, string) {
	if param.year == "" {
		return false, "You haven't specified a search year"
	} else if len(param.months) == 0 && param.lineLimit == 0 {
		return false, "You haven't specified a search month, or a line limit number!"
	}
	return true, "Ready!"
}

func getCommentDataFromLine(line []byte, keyTypes map[string]string) map[string]string {
	result := make(map[string]string, 0)

	for key, v := range keyTypes {
		if v == "str" {
			result[key] = strings.ToLower(fastjson.GetString(line, key))
		} else if v == "int" {
			result[key] = strconv.Itoa(fastjson.GetInt(line, key))
		} else {
			log.Fatal("Undetected type: " + v)
		}
	}
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

func SaveCriteriaDataToFile(criteria string, value string, year string, basedir string, schema commentSchema) {
	for _, v := range AllMonths {
		relevantComments := make([]map[string]string, 0)
		str := basedir + "/" + v + "/" + criteria + "_" + value + "_" + schema.name
		if _, err := os.Stat(str); !os.IsNotExist(err) {
			fmt.Println("Found cached data for " + str)
			continue
		}
		var buffer bytes.Buffer
		buffer.Write([]byte(basedir))
		buffer.Write([]byte("/RC_" + year + monthToIntString(v)))
		file, fileOpenErr := os.Open(buffer.String())
		metafile, fileOpenErr2 := os.Open(buffer.String() + "_meta.txt")

		if fileOpenErr != nil {
			fmt.Print(fileOpenErr)
			os.Exit(0)
		}
		var totalLines uint64

		if fileOpenErr2 != nil {
			totalLines = 0
			metafile.Close()
		} else {
			rdr := bufio.NewReader(metafile)
			line, _, _ := rdr.ReadLine()
			totalLines, _ = strconv.ParseUint(string(line), 10, 64)
		}

		reader := bufio.NewReaderSize(file, 4096)

		fmt.Println("Opened file " + buffer.String())

		var linesRead uint64 = 0

		startTime := time.Now()

		for {
			line := recurseBuildCompleteLine(reader)
			if line == nil {
				fmt.Println("Lines: " + strconv.FormatUint(linesRead, 10))
				log.Println("Encountered error; concluding analysis")
				break
			}
			linesRead++

			if strings.Contains(string(line), "\""+criteria+"\":\""+value+"\"") {
				parsed := getCommentDataFromLine(line, schema.schema)
				relevantComments = append(relevantComments, parsed)
			}

			if linesRead%HundredThousand == 0 {
				if totalLines != 0 {
					percentDone := (float64(linesRead) / float64(totalLines)) * 100.0
					fmt.Println("Percent complete with " + v + ": " + strconv.FormatFloat(percentDone, 'f', 2, 64) + "%")
				} else {
					fmt.Println("Lines complete: " + strconv.FormatUint(linesRead, 10))
				}

			}
		}
		dif := time.Now().Sub(startTime).String()
		fmt.Println("Took " + dif + " to search " + strconv.FormatUint(linesRead, 10) + " comments of file " + buffer.String())
		dumpDataToFilepath(relevantComments, str)
	}
}

func OpenExtractedDatafile(basedir string, subreddit string, extractedType string) {
	var commentData []map[string]string

	for _, v := range AllMonths {
		fmt.Println("Reading " + v)
		var tempCommData []map[string]string
		str := basedir + "/" + v + "/" + subreddit + "_" + extractedType
		plan, fileOpenErr := ioutil.ReadFile(str)
		if fileOpenErr != nil {
			log.Fatal("failed to open " + str)
		}
		err := json.Unmarshal(plan, &tempCommData)
		if err != nil {
			log.Println(err)
		}
		commentData = append(commentData, tempCommData...)
	}
	fmt.Println(strconv.Itoa(len(commentData)) + " total comments")

	fmt.Println("Tallying word counts...")
	tallies := tallyWordOccurrences(commentData)
	for _, v := range tallies[:10] {
		percent := (float64(v.Value) / float64(len(commentData))) * 100.0
		fmt.Println(v.Key + ": " + strconv.FormatUint(v.Value, 10) + " comment occurrences (" +
			strconv.FormatFloat(percent, 'f', 3, 64) + "%)")
	}

	keys := make([]string, 20)
	i := 0
	for _, v := range tallies[:20] {
		keys[i] = v.Key
		i++
	}
	fmt.Println("Tallying karma counts...")
	karmas := tallyWordKarma(keys, commentData)
	for _, karma := range karmas[:10] {
		for _, pair := range tallies {
			if pair.Key == karma.Key {
				karmaPerOccurence := float64(karma.Value) / float64(pair.Value)
				fmt.Println(karma.Key + ": " + strconv.FormatFloat(karmaPerOccurence, 'f', 5, 64) + " karma per comment")
				continue
			}
		}
	}
}

func rankByWordCount(wordFrequencies map[string]uint64) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string
	Value uint64
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func tallyWordOccurrences(comments []map[string]string) PairList {
	tallies := make(map[string]uint64, 0)
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
			if _, ok := wordPresence[word]; !ok {
				wordPresence[word] = true
				if _, ok := tallies[word]; ok {
					tallies[word]++
				} else {
					tallies[word] = 1
				}
			}
		}
	}
	return rankByWordCount(tallies)
}

func tallyWordKarma(words []string, comments []map[string]string) PairList {
	wordKarmas := make(map[string]uint64)
	filter, _ := regexp.Compile("[^a-zA-Z0-9]+")
	for _, comment := range comments {
		subwords := strings.Fields(comment["body"])
		for _, word := range words {
			for _, subword := range subwords {
				subword = strings.ToLower(subword)
				subword = filter.ReplaceAllString(subword, "")
				if word == subword {
					karma, err := strconv.ParseInt(comment["score"], 10, 64)
					if err != nil {
						log.Fatal(err)
					}
					if _, ok := wordKarmas[word]; ok {
						wordKarmas[word] += uint64(karma)
					} else {
						wordKarmas[word] = uint64(karma)
					}
					break
				}
			}
		}
	}
	return rankByWordCount(wordKarmas)
}

func dumpDataToFilepath(data []map[string]string, filePath string) {
	f, err := os.Create(filePath)
	if err != nil {
		log.Println(err)
	}

	bytess, err2 := json.Marshal(data)
	if err2 != nil {
		log.Println(err2)
	}
	f.Write(bytess)
	f.Close()
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
