package selection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
)

type SubredditSplitScores struct {
	A_Scores     []int64
	B_Scores     []int64
	Total_Scores []int64 //what all the comments look like for a user
}

//NOTE changing this datamodel can invalidate older or future data captures...

var commentFields = map[string]string{
	"author":    "str",
	"subreddit": "str",
	"body":      "str",
	"score":     "int",
}

var BasicSchema = commentSchema{
	name:   "Basic",
	schema: commentFields}

type commentSchema struct {
	name   string
	schema map[string]string
}

func OpenCachedOrProcessAndFilterMonth(searchCriteria *SearchParams, month string, BaseDir string, suffix string) []map[string]string {
	var commentData []map[string]string
	searchCriteria.months = []string{month} //make sure the saved file only refers to this month specifically
	res := searchCriteria.ValuesToString()
	dir := BaseDir + "/" + month + "/" + res + suffix
	fmt.Println(dir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("Cached data not found for these specs; reanalyzing")
		commentData = FilterAllMonthsComments(searchCriteria, BaseDir, suffix)

		fmt.Println("Total comments matching criteria: " + strconv.Itoa(len(commentData)))
		if len(commentData) < 1000 {
			for _, v := range commentData {
				fmt.Println(v)
			}
		}
		f, err := os.Create(dir)
		defer f.Close()
		if err != nil {
			log.Println(err)
		}

		bytes, err2 := json.Marshal(commentData)
		if err2 != nil {
			log.Println(err2)
		}
		f.Write(bytes)
	} else {
		fmt.Println("Previous data found for this spec!")
		plan, _ := ioutil.ReadFile(dir)
		err := json.Unmarshal(plan, &commentData)
		if err != nil {
			fmt.Println(err)
		}
	}
	return commentData
}

func FilterAllMonthsComments(searchCriteria *SearchParams, baseDir, suffix string) []map[string]string {
	allMonthsComments := make([]map[string]string, 0)

	for _, v := range searchCriteria.months {
		allMonthsComments = append(allMonthsComments, OpenCachedOrProcessAndFilterMonth(searchCriteria, v, baseDir, suffix)...)
	}

	return allMonthsComments
}

type subAndFloat struct {
	subreddit string
	percent   float64
}

type byLength []subAndFloat

func (s byLength) Len() int {
	return len(s)
}

func (s byLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byLength) Less(i, j int) bool {
	return s[i].percent > s[j].percent
}

func SubredditStats(sub string, basedir string) string {
	ANDList := []string{"\"subreddit\":\"" + sub + "\""}

	searchCriteria := MakeSimpleSearchParams("2016", []string{"Jan", "Feb"}, 0, []string{}, ANDList)
	data := FilterAllMonthsComments(searchCriteria, basedir, "")
	totalComments := len(data)
	var totalScore uint64 = 0

	for _, v := range data {
		num, err := strconv.ParseInt(v["score"], 10, 64)
		if err != nil {
			log.Println(err)
		}
		totalScore += uint64(num)
	}
	avgScore := float64(totalScore) / float64(totalComments)

	avgAdjusted := 0.0
	for _, v := range data {
		num, _ := strconv.ParseInt(v["score"], 10, 64)
		sqred := (float64(num) - avgScore) * (float64(num) - avgScore)
		avgAdjusted += sqred
	}
	stdDev := math.Sqrt(avgAdjusted / float64(totalComments))

	return sub + "," + strconv.FormatFloat(avgScore, 'f', 3, 64) + "," + strconv.FormatFloat(stdDev, 'f', 3, 64)
}

func AuthorSubredditStats(author string, baseDir string) string {
	ANDList := []string{"\"author\":\"" + author + "\""}

	searchCriteria := MakeSimpleSearchParams("2016", AllMonths, 0, []string{}, ANDList)
	data := FilterAllMonthsComments(searchCriteria, baseDir, "allSubreddits")

	totalComments := len(data)
	subredditScores := make(map[string][]int64, 0)
	totalScores := make([]int64, totalComments)

	for _, v := range data {
		if val, ok := subredditScores[v["subreddit"]]; ok {
			num, err := strconv.ParseInt(v["score"], 10, 64)
			if err != nil {
				log.Println(err)
			}
			subredditScores[v["subreddit"]] = append(val, num)
		} else {
			subredditScores[v["subreddit"]] = make([]int64, 1)
			num, err := strconv.ParseInt(v["score"], 10, 64)
			if err != nil {
				log.Println(err)
			}
			subredditScores[v["subreddit"]][0] = num
		}
		num, err := strconv.ParseInt(v["score"], 10, 64)
		if err != nil {
			log.Println(err)
		}
		totalScores = append(totalScores, num)
	}
	var buffer bytes.Buffer
	buffer.Write([]byte("Total comments: " + strconv.Itoa(totalComments) + "\n"))

	commentRatios := make([]subAndFloat, 0)
	scoreRatios := make([]subAndFloat, 0)

	var totalSum int64 = 0

	for _, v := range totalScores {
		totalSum += v
	}
	for key, sub := range subredditScores {
		var subredditSum float64 = 0

		pair := subAndFloat{key, (float64(len(sub)) / float64(totalComments)) * 100.0}
		commentRatios = append(commentRatios, pair)
		for _, v := range sub {
			subredditSum += float64(v)
		}
		pair2 := subAndFloat{key, (float64(subredditSum) / float64(totalSum)) * 100.0}
		scoreRatios = append(scoreRatios, pair2)
	}
	sort.Sort(byLength(commentRatios))
	sort.Sort(byLength(scoreRatios))

	for _, v := range commentRatios {
		for _, vv := range scoreRatios {
			if vv.subreddit == v.subreddit {
				buffer.WriteString(v.subreddit + "," + strconv.FormatFloat(v.percent, 'f', 3, 64) + "," + strconv.FormatFloat(vv.percent, 'f', 3, 64) + "\n")
			}
		}
	}
	return buffer.String()
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
