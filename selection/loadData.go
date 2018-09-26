package selection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	"gilded":    "int",
}

func OpenCachedOrProcessAndFilterMonth(searchCriteria *SearchParams, month string, BaseDir string, suffix string) []map[string]string {
	var commentData []map[string]string
	searchCriteria.months = []string{month} //make sure the saved file only refers to this month specifically
	res := searchCriteria.ValuesToString()
	dir := BaseDir + "/" + month + "/" + res + suffix
	fmt.Println(dir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("Cached data not found for these specs; reanalyzing")
		commentData = OpenAndSearchFile(*searchCriteria, BaseDir, commentFields)

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
	buffer.Write([]byte("Total comments: " + strconv.Itoa(totalComments)))

	var totalSum int64 = 0

	for _, v := range totalScores {
		totalSum += v
	}
	for key, sub := range subredditScores {
		var subredditSum int64 = 0
		buffer.Write([]byte("Subreddit ratio: " + strconv.FormatFloat(float64(len(sub))/float64(totalComments), 'f', 3, 64)))
		for _, v := range sub {
			subredditSum += v
		}
		buffer.Write([]byte("Subreddit " + key + " score ratio: " + strconv.FormatFloat(float64(subredditSum)/float64(totalSum), 'f', 3, 64)))
	}
	return buffer.String()
}

func CompareSubreddits(allComments []map[string]string, subA string, subB string) {
	var authorsInEither = make(map[string]SubredditSplitScores, 0)
	for i, comment := range allComments { //individual comment
		author := comment["author"]
		if _, ok := authorsInEither[author]; !ok {
			authorsInEither[author] = SubredditSplitScores{make([]int64, 0), make([]int64, 0), make([]int64, 0)}
		}
		if i%100000 == 0 {
			fmt.Println("Done checking " + strconv.Itoa(i) + " comments")
		}
		splitScore := authorsInEither[author]
		score, err := strconv.ParseInt(comment["score"], 10, 64)

		if err != nil {
			fmt.Println(err)
		}

		if comment["subreddit"] == subA {
			splitScore.A_Scores = append(splitScore.A_Scores, score)
		} else if comment["subreddit"] == subB {
			splitScore.B_Scores = append(splitScore.B_Scores, score)
		}
		authorsInEither[author] = splitScore
	}
	var total = 0
	var aAvgs = make([]float64, 0)
	var bAvgs = make([]float64, 0)
	var threshold = 20
	fmt.Println("Only including users with more than " + strconv.Itoa(threshold) + " posts in each subreddit")
	for _, v := range authorsInEither {
		if len(v.A_Scores) > threshold && len(v.B_Scores) > threshold { //comment to both...
			aAvg := average(v.A_Scores)
			bAvg := average(v.B_Scores)
			//get the average of all comments for a user in each subreddit

			aAvgs = append(aAvgs, aAvg)
			bAvgs = append(bAvgs, bAvg)
			total++
		}
	}
	fmt.Println("Total of " + strconv.Itoa(total) + " users post in both channels")
	cumaAvg := faverage(aAvgs)
	cumbAvg := faverage(bAvgs)
	fmt.Println("The average score for " + subA + " is " + strconv.FormatFloat(cumaAvg, 'f', 4, 64))
	fmt.Println("The average score for " + subB + " is " + strconv.FormatFloat(cumbAvg, 'f', 4, 64))
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
