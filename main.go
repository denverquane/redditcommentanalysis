package main

import (
	"fmt"
	"github.com/denverquane/redditcommentanalysis/selection"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

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
		selection.OpenExtractedDatafile(os.Getenv("BASE_DATA_DIRECTORY")+"/"+year, subreddit, schema)
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

	muxRouter.HandleFunc("/extractSub/{subreddit}", handleExtractSub).Methods("GET")

	muxRouter.HandleFunc("/viewSub/{subreddit}", handleViewSub).Methods("GET")
	return muxRouter
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

		go processQueue()
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
}

func processQueue() {
	tempSub := ""
	for len(extractSubQueue) > 0 {
		tempSub = extractSubQueue[0]
		selection.SaveCriteriaDataToFile("subreddit", tempSub, "2016",
			os.Getenv("BASE_DATA_DIRECTORY"), selection.BasicSchema)
		extractSubQueue = extractSubQueue[1:] //done
		fmt.Println("COMPLETED")
	}
}

func handleViewSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subreddit := vars["subreddit"]

	go selection.OpenExtractedDatafile(os.Getenv("BASE_DATA_DIRECTORY")+"/"+"2016", subreddit, "Basic")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
}
