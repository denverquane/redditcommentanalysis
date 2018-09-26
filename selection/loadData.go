package selection

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type SubredditSplitScores struct {
	A_Scores []int64
	B_Scores []int64
}

var commentFields = map[string]string{
	"author":    "str",
	"subreddit": "str",
	//"body":      "str",
	"score":  "int",
	"gilded": "int",
}

func FilterMonthWorker(criteria *SearchParams, month string, baseDir string, c chan []map[string]string) {
	log.Println("Started worker for: " + month)
	data := OpenCachedOrProcessAndFilterMonth(criteria, month, baseDir)
	c <- data
	log.Println(month + " worker finished")
}

func OpenCachedOrProcessAndFilterMonth(searchCriteria *SearchParams, month string, BaseDir string) []map[string]string {
	var commentData []map[string]string
	searchCriteria.months = []string{month} //make sure the saved file only refers to this month specifically
	res := searchCriteria.ValuesToString()
	dir := BaseDir + "/" + month + "/" + res
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

func FilterAllMonthsComments(searchCriteria *SearchParams, baseDir string) []map[string]string {
	allMonthsComments := make([]map[string]string, 0)
	totalMonths := len(searchCriteria.months)

	ch := make(chan []map[string]string, totalMonths)
	for _, v := range searchCriteria.months {
		go FilterMonthWorker(searchCriteria, v, baseDir, ch)
	}

	for i := 0; i < totalMonths; i++ {
		allMonthsComments = append(allMonthsComments, <-ch...)
	}

	return allMonthsComments
}

func CompareSubreddits(allComments []map[string]string, subA string, subB string) {
	var authorsInEither = make(map[string]SubredditSplitScores, 0)
	for i, comment := range allComments { //individual comment
		author := comment["author"]
		if _, ok := authorsInEither[author]; !ok {
			authorsInEither[author] = SubredditSplitScores{make([]int64, 0), make([]int64, 0)}
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
	fmt.Println("Only including users with more than 10 posts in each subreddit")
	for _, v := range authorsInEither {
		if len(v.A_Scores) > 1 && len(v.B_Scores) > 1 { //comment to both...
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
