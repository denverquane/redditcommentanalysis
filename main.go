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
var subredditStatuses = make(map[string]subredditStatus)

type subredditStatus struct {
	extracting       bool
	extractedSummary string
	processing       bool
	processedSummary string
}

func (status subredditStatus) ToString() string {
	str := "Extracting: " + strconv.FormatBool(status.extracting) + "\n"

	if !status.extracting && status.extractedSummary != "" {
		str += "Extracted summary: " + status.extractedSummary + "\n"
	}

	str += "Processing: " + strconv.FormatBool(status.processing) + "\n"

	if !status.processing && status.processedSummary != "" {
		str += "Processed summary: " + status.processedSummary + "\n"
	}

	return str
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	checkForExtractedSubs("2016", "Basic")

	if os.Getenv("RUN_SERVER") == "true" {
		port := os.Getenv("SERVER_PORT")
		log.Fatal(run(port))
	} else {
		//year := "2016"
		//subreddit := "funny"
		//schema := "Basic"
		var prog float64
		_ = selection.SaveCriteriaDataToFile("subreddit", "funny", "2016", os.Getenv("BASE_DATA_DIRECTORY"), selection.BasicSchema, &prog)
		//selection.OpenExtractedDatafile(os.Getenv("BASE_DATA_DIRECTORY")+"/"+year, subreddit, schema)
		//scanDirForExtractedSubData(os.Getenv("BASE_DATA_DIRECTORY") + "/2016/Jan", "Basic")

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

func checkForExtractedSubs(year string, schema string) {
	dir := os.Getenv("BASE_DATA_DIRECTORY") + "/" + year + "/"
	arr := selection.ScanDirForExtractedSubData(dir+"Jan", schema)
	for _, sub := range arr {

		//TODO actually insert status (total comments?) for a given subreddit
		subredditStatuses[sub] = subredditStatus{extracting: false, extractedSummary: "gh", processing: false, processedSummary: ""}
	}

	for _, v := range selection.AllMonths {
		arr := selection.ScanDirForExtractedSubData(dir+v, schema)
		for prevSubName := range subredditStatuses {
			found := false
			for _, newSubName := range arr {
				if newSubName == prevSubName {
					found = true
					break
				}
			}
			if !found {
				fmt.Println("Missing successive months for " + prevSubName)
			} else {
				fmt.Println("All months successfully validated for " + prevSubName)
			}
		}
	}
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()

	muxRouter.HandleFunc("/", handleIndex).Methods("GET")
	muxRouter.HandleFunc("/status", handleStatus).Methods("GET")
	muxRouter.HandleFunc("/extractSub/{subreddit}", handleExtractSub).Methods("GET")
	muxRouter.HandleFunc("/status/{subreddit}", handleViewStatus).Methods("GET")
	muxRouter.HandleFunc("/processSub/{subreddit}", handleProcessSub).Methods("GET")
	return muxRouter
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Welcome to the Reddit Comment Extractor!\nThese are the endpoints to use:\n")
	io.WriteString(w, "GET \"/status\" displays the overall status of this server and its data processing\n")
	io.WriteString(w, "GET \"/extractSub/<sub>\" extracts ALL comments from the <sub> subreddit, and saves to a datafile for later processing\n")
	io.WriteString(w, "GET \"/processSub/<sub>\" processes the previously-extracted data for a subreddit, and saves these processed analytics for later retrieval\n")
	io.WriteString(w, "GET \"/status/<sub>\" displays the extraction and processing status for a subreddit\n")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	str := ""
	if len(extractSubQueue) > 0 {
		str += "These subreddits are in the queue and waiting to be extracted: \n"
		for _, v := range extractSubQueue {
			str += v + "\n"
			if vv, ok := subredditStatuses[v]; ok {
				if vv.extracting {
					str += " (currently extracting, " + strconv.FormatFloat(extractingProg, 'f', 2, 64) + "% complete)\n"
				}
			}
		}
	} else {
		str += "Subreddit processing queue is empty: waiting for jobs!"
	}
	io.WriteString(w, str)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
}

func handleExtractSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subreddit := vars["subreddit"]

	if val, ok := subredditStatuses[subreddit]; ok {
		if val.extracting || val.extractedSummary != "" {
			io.WriteString(w, "Subreddit "+subreddit+" has been extracted, is extracting, or is in the queue for extraction!\n")
			io.WriteString(w, val.ToString())
		}
	} else {
		extractSubQueue = append(extractSubQueue, subreddit)
		io.WriteString(w, subreddit+" appended to queue!\nCurrent length: "+
			strconv.Itoa(len(extractSubQueue))+"(Including the job currently processing)")
		subredditStatuses[subreddit] = subredditStatus{}

		if len(extractSubQueue) == 1 {
			go extractQueue() //this is the main goroutine that will process all the future jobs
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
}

var extractingProg float64

func extractQueue() {
	tempSub := ""
	for len(extractSubQueue) > 0 {
		tempSub = extractSubQueue[0]
		if v, ok := subredditStatuses[tempSub]; !ok {
			subredditStatuses[tempSub] = subredditStatus{extracting: true, extractedSummary: "", processing: false, processedSummary: ""}
		} else {
			v.extracting = true
			subredditStatuses[tempSub] = v
		}
		str := selection.SaveCriteriaDataToFile("subreddit", tempSub, "2016",
			os.Getenv("BASE_DATA_DIRECTORY"), selection.BasicSchema, &extractingProg)

		v := subredditStatuses[tempSub]
		v.extracting = false
		v.extractedSummary = str
		subredditStatuses[tempSub] = v
		extractSubQueue = extractSubQueue[1:] //done
		fmt.Println("COMPLETED")
	}
}

func handleProcessSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subreddit := vars["subreddit"]

	if val, ok := subredditStatuses[subreddit]; ok {
		if val.processing {
			io.WriteString(w, "Subreddit \""+subreddit+"\" is still being processed!")
		} else if val.processedSummary != "" {
			io.WriteString(w, "Subreddit \""+subreddit+"\" has already been processed:\n"+val.extractedSummary)
		} else if val.extracting {
			io.WriteString(w, "Subreddit \""+subreddit+"\" is still being extracted")
		} else if val.extractedSummary != "" {
			go selection.OpenExtractedSubredditDatafile(os.Getenv("BASE_DATA_DIRECTORY")+"/"+"2016", subreddit, "Basic")
		} else {
			io.WriteString(w, "Subreddit has not been extracted or processed yet, please hit the endpoint \"/extractSub/"+subreddit+"\" to extract the subreddit data first")
		}
	} else {
		io.WriteString(w, "Subreddit has not been extracted or processed yet, please hit the endpoint \"/extractSub/"+subreddit+"\" to extract the subreddit data first")
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
}

func handleViewStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subreddit := vars["subreddit"]

	if val, ok := subredditStatuses[subreddit]; !ok {
		io.WriteString(w, "Subreddit has not been extracted or processed yet!")
	} else {
		io.WriteString(w, val.ToString())
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
}
