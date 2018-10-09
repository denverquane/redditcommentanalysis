package main

import (
	"fmt"
	"github.com/denverquane/redditcommentanalysis/selection"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/joho/godotenv"
	"os"
)

var authorStatusMap = make(map[string]string, 0)
var subredditStatusMap = make(map[string]string, 0)
var extractSubQueue = make([]string, 0)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}


	if os.Getenv("RUN_SERVER") == "true" {
		port := os.Getenv("SERVER_PORT")
		log.Fatal(run(port))
	} else {
		year := "2016"
		subreddit := "funny"
		schema := "Basic"
		//searchCriteria := selection.MakeSimpleSearchParams("2016", []string{"Dec"}, 0,
		//	[]string{}, []string{"\"subreddit\":\"" + "pics" + "\""})
		//_ = selection.FilterAllMonthsComments(searchCriteria, BaseDataDirectory, "")
		//fmt.Println(allMonths)
		selection.OpenExtractedDatafile(os.Getenv("BASE_DATA_DIRECTORY") + "/" + year, subreddit, schema)
	}
}

func run(port string) error {
	handler := makeMuxRouter()

	s := &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()

	muxRouter.HandleFunc("/runAuthorAnalysis/{author}", handleRunAuthor).Methods("GET")

	muxRouter.HandleFunc("/extractSub/{subreddit}", handleExtractSub).Methods("GET")
	muxRouter.HandleFunc("/extractSub/status", handleExtractSubStatus).Methods("GET")
	muxRouter.HandleFunc("/runSubredditAnalysis/{subreddit}", handleRunSubreddit).Methods("GET")

	muxRouter.HandleFunc("/authorStatus/{author}", handleGetAuthorStatus).Methods("GET")
	muxRouter.HandleFunc("/subStatus/{sub}", handleGetSubredditStatus).Methods("GET")

	return muxRouter
}

func handleGetAuthorStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	str := authorStatusMap[author]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	io.WriteString(w, str)
}

func handleExtractSubStatus(w http.ResponseWriter, r *http.Request) {
	str := ""
	if len(extractSubQueue) > 0 {
		str = "Currently processing " + extractSubQueue[0] + ", " + strconv.Itoa(len(extractSubQueue)-1) + " jobs remaining"
	} else {
		str = "No pending jobs!"
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	io.WriteString(w, str)
}

func handleExtractSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subreddit := vars["subreddit"]

	if len(extractSubQueue) > 0 {
		extractSubQueue = append(extractSubQueue, subreddit)
		io.WriteString(w, subreddit+" appended to queue!\nCurrent length: "+
			strconv.Itoa(len(extractSubQueue))+"(Including the job currently processing)")
	} else {
		extractSubQueue = append(extractSubQueue, subreddit)
		io.WriteString(w, subreddit+" appended to queue!\nCurrent length: "+
			strconv.Itoa(len(extractSubQueue))+"(Including the job currently processing)")
		tempSub := ""
		for len(extractSubQueue) > 0 {
			tempSub = extractSubQueue[0]
			selection.SaveCriteriaDataToFile("subreddit", tempSub, "2016",
				os.Getenv("BASE_DATA_DIRECTORY"), selection.BasicSchema)
			extractSubQueue = extractSubQueue[1:] //done
			fmt.Println("COMPLETED")
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
}

func handleGetSubredditStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sub := vars["sub"]
	str := subredditStatusMap[sub]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	io.WriteString(w, str)
}

func handleRunAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	fmt.Println("Start run")
	//PendingSearchCriteria = selection.InterpretFromHttp(r)
	var text string
	//ready, _ := PendingSearchCriteria.IsReady()
	//if !ready {
	//	fmt.Println("Not ready yet")
	//	text = "Not ready yet:\n" + PendingSearchCriteria.ToString()
	//} else {
	if _, ok := authorStatusMap[author]; ok {
		text = "Already running, or has been ran!!"
	} else {
		text = "Running for " + author
		go waitForFilter(author)
		//allMonths := selection.FilterAllMonthsComments(PendingSearchCriteria, BaseDataDirectory, "")
		////selection.CompareSubreddits(allMonths, SubA, SubB)
		//}
	}
	w.WriteHeader(http.StatusOK)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	io.WriteString(w, string(text))
}

func handleRunSubreddit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sub := vars["subreddit"]
	fmt.Println("Start run")
	//PendingSearchCriteria = selection.InterpretFromHttp(r)
	var text string

	go waitForChannelStats(sub)
	//allMonths := selection.FilterAllMonthsComments(PendingSearchCriteria, BaseDataDirectory, "")
	////selection.CompareSubreddits(allMonths, SubA, SubB)
	//}
	w.WriteHeader(http.StatusOK)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	io.WriteString(w, string(text))
}

func waitForChannelStats(sub string) {
	subredditStatusMap[sub] = "Running"
	str := selection.SubredditStats(sub, os.Getenv("BASE_DATA_DIRECTORY"))
	subredditStatusMap[sub] = str
}

func waitForFilter(author string) {
	authorStatusMap[author] = "Running"
	str := selection.AuthorSubredditStats(author, os.Getenv("BASE_DATA_DIRECTORY"))
	authorStatusMap[author] = str
}
