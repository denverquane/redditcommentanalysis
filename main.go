package main

import (
	"fmt"
	"github.com/denverquane/redditcommentanalysis/selection"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"time"
)

const BaseDataDirectory = "D:/Reddit_Data"
const SSDBaseDataDirectory = "C:/Users/Denver"

const RunServer = false

var PendingSearchCriteria = selection.MakeEmptySearchParams()

const SubA = "dogs"
const SubB = "aww"

var authorStatusMap = make(map[string]string, 0)
var subredditStatusMap = make(map[string]string, 0)

func main() {

	if RunServer {
		log.Fatal(run())
	} else {
		//begin := time.Now()
		searchCriteria := selection.MakeSimpleSearchParams("2016", []string{"Feb"}, 0,
			[]string{}, []string{"\"author\":\"" + "spez" + "\"", "\"subreddit\":\"" + "the_donald" + "\""})
		allMonths := selection.FilterAllMonthsComments(searchCriteria, BaseDataDirectory, "")
		fmt.Println(allMonths)
		//selection.CompareSubreddits(allMonths, SubA, SubB)

		//str := selection.AuthorSubredditStats("gallowboob", BaseDataDirectory)
		//fmt.Println(str)

		//total := time.Now().Sub(begin).String()
		//fmt.Println("Took " + total + " for " + searchCriteria.ToString())
	}
}

func run() error {
	handler := makeMuxRouter()

	s := &http.Server{
		Addr:           ":5050",
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

	//muxRouter.HandleFunc("/", handleGetStatus).Methods("GET")

	muxRouter.HandleFunc("/runAuthorAnalysis/{author}", handleRunAuthor).Methods("GET")

	muxRouter.HandleFunc("/extractSub/{subreddit}", handleExtractSub).Methods("GET")
	muxRouter.HandleFunc("/runSubredditAnalysis/{subreddit}", handleRunSubreddit).Methods("GET")

	muxRouter.HandleFunc("/authorStatus/{author}", handleGetAuthorStatus).Methods("GET")
	muxRouter.HandleFunc("/subStatus/{sub}", handleGetSubredditStatus).Methods("GET")

	return muxRouter
}

//
//func handleGetStatus(w http.ResponseWriter, r *http.Request) {
//	w.WriteHeader(http.StatusOK)
//
//	fmt.Println("GET status")
//	w.Header().Set("Access-Control-Allow-Origin", "*")
//	w.Header().Add("Access-Control-Allow-Methods", "PUT")
//	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
//	io.WriteString(w, PendingSearchCriteria.ToString())
//}

func handleGetAuthorStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	str := authorStatusMap[author]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	io.WriteString(w, str)
}

func handleExtractSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subreddit := vars["subreddit"]
	selection.SaveSubredditDataToFile(subreddit, "2016", BaseDataDirectory)
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
	str := selection.SubredditStats(sub, BaseDataDirectory)
	subredditStatusMap[sub] = str
}

func waitForFilter(author string) {
	authorStatusMap[author] = "Running"
	str := selection.AuthorSubredditStats(author, BaseDataDirectory)
	authorStatusMap[author] = str
}
