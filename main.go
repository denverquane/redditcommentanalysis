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

func main() {

	if RunServer {
		log.Fatal(run())
	} else {
		//begin := time.Now()
		//searchCriteria := selection.MakeSimpleSearchParams("2016", []string{"Jan"}, 0,
		//	[]string{"\"subreddit\":\"" + SubA + "\"", "\"subreddit\":\"" + SubB + "\""}, []string{})
		//allMonths := selection.FilterAllMonthsComments(searchCriteria, BaseDataDirectory)
		//selection.CompareSubreddits(allMonths, SubA, SubB)
		selection.AuthorSubredditStats("mrsoupsox", BaseDataDirectory)

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

	muxRouter.HandleFunc("/setCriteria", handleSetCriteria).Methods("POST")

	muxRouter.HandleFunc("/run", handleRunCommand).Methods("GET")

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

func handleSetCriteria(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Set criteria")
	PendingSearchCriteria = selection.InterpretFromHttp(r)
	defer r.Body.Close()

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	//io.WriteString(w, string(text))
}

func handleRunCommand(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST run")
	var text string
	ready, _ := PendingSearchCriteria.IsReady()
	if !ready {
		fmt.Println("Not ready yet")
		text = "Not ready yet:\n" + PendingSearchCriteria.ToString()
	} else {
		text = "Running: \n" + PendingSearchCriteria.ToString()
		allMonths := selection.FilterAllMonthsComments(PendingSearchCriteria, BaseDataDirectory)
		selection.CompareSubreddits(allMonths, SubA, SubB)
	}

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	io.WriteString(w, string(text))
}
