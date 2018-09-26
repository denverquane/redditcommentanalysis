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

const RunServer = true

var PendingSearchCriteria = selection.MakeEmptySearchParams()

const SubA = "dogs"
const SubB = "aww"

var authorStatusMap = make(map[string]string, 0)

func main() {

	if RunServer {
		log.Fatal(run())
	} else {
		//begin := time.Now()
		//searchCriteria := selection.MakeSimpleSearchParams("2016", []string{"Jan"}, 0,
		//	[]string{"\"subreddit\":\"" + SubA + "\"", "\"subreddit\":\"" + SubB + "\""}, []string{})
		//allMonths := selection.FilterAllMonthsComments(searchCriteria, BaseDataDirectory)
		//selection.CompareSubreddits(allMonths, SubA, SubB)
		str := selection.AuthorSubredditStats("gallowboob", BaseDataDirectory)
		fmt.Println(str)

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

	muxRouter.HandleFunc("/", handleGetStatus).Methods("GET")

	muxRouter.HandleFunc("/runAuthorAnalysis/{author}", handleRunCommand).Methods("GET")

	muxRouter.HandleFunc("/status/{id}", handleGetIDStatus).Methods("GET")

	return muxRouter
}

func handleGetStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	fmt.Println("GET status")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	io.WriteString(w, PendingSearchCriteria.ToString())
}

func handleGetIDStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	str := authorStatusMap[id]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	io.WriteString(w, str)
}

func handleRunCommand(w http.ResponseWriter, r *http.Request) {
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
		text = "Running: \n" + PendingSearchCriteria.ToString()
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

func waitForFilter(author string) {
	authorStatusMap[author] = "Running"
	str := selection.AuthorSubredditStats(author, BaseDataDirectory)
	authorStatusMap[author] = str
}
