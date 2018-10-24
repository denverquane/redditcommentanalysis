package main

import (
	"encoding/json"
	"fmt"
	"github.com/denverquane/redditcommentanalysis/selection"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

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

	for _, month := range selection.AllMonths {
		arr := selection.ScanDirForExtractedSubData(dir+month, schema) //scan the current month folder for all subs

		for _, sub := range arr {
			fmt.Println("Checking " + sub)
			var commentCount int64
			str := os.Getenv("BASE_DATA_DIRECTORY") + "/" + year + "/" + month + "/" + "subreddit_" + sub + "_" + schema + "_count"
			plan, fileOpenErr := ioutil.ReadFile(str)
			if fileOpenErr != nil {
				log.Println("Failed to open " + str + ", now opening datafile and writing total comment count to file")
				var comments []map[string]string
				str := os.Getenv("BASE_DATA_DIRECTORY") + "/" + year + "/" + month + "/" + "subreddit_" + sub + "_" + schema
				plan, fileOpenErr := ioutil.ReadFile(str)
				if fileOpenErr != nil {
					log.Fatal("failed to open " + str)
				}
				err := json.Unmarshal(plan, &comments)
				if err != nil {
					log.Println(err)
				}
				file, err2 := os.Create(os.Getenv("BASE_DATA_DIRECTORY") + "/" + year + "/" + month + "/" + "subreddit_" + sub + "_" + schema + "_count")
				if err != nil {
					log.Println(err2)
				}
				length := strconv.FormatInt(int64(len(comments)), 10)
				file.Write([]byte(length))
				fmt.Println("Wrote " + length + " to file for " + sub + " in " + month)

				commentCount = int64(len(comments))
				file.Close()
			} else {
				err := json.Unmarshal(plan, &commentCount)
				if err != nil {
					log.Fatal("Failed to unmarshal extracted data for " + sub)
				} else {
					fmt.Println("Extracted " + strconv.FormatInt(commentCount, 10))
				}
			}

			if val, ok := subredditStatuses[sub]; ok {
				val.ExtractedMonthCommentCounts[month] = commentCount
			} else {
				status := subredditStatus{Extracting: false, ExtractedMonthCommentCounts: make(map[string]int64, 1), Processing: false, ProcessedSummary: ""}
				status.ExtractedMonthCommentCounts[month] = commentCount
				subredditStatuses[sub] = status
			}
		}
	}
	for sub := range subredditStatuses {
		fmt.Println("Found all month's entries for " + sub)
	}
}

var extractSubQueue = make([]string, 0)
var processSubQueue = make([]string, 0)
var subredditStatuses = make(map[string]subredditStatus)

type subredditStatus struct {
	Extracting                  bool
	ExtractedMonthCommentCounts map[string]int64
	Processing                  bool
	ProcessedSummary            string
}

func (status subredditStatus) ToString() string {
	str := "Extracting: " + strconv.FormatBool(status.Extracting) + "\n"

	if !status.Extracting && len(status.ExtractedMonthCommentCounts) != 0 {
		str += "Extracted summary: " + "\n"
	}

	str += "Processing: " + strconv.FormatBool(status.Processing) + "\n"

	if !status.Processing && status.ProcessedSummary != "" {
		str += "Processed summary: " + status.ProcessedSummary + "\n"
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
		year := "2016"
		subreddit := "funny"
		schema := "Basic"
		//var prog float64
		//_ = selection.SaveCriteriaDataToFile("subreddit", "funny", "2016", os.Getenv("BASE_DATA_DIRECTORY"), selection.BasicSchema, &prog)
		selection.OpenExtractedSubredditDatafile(os.Getenv("BASE_DATA_DIRECTORY")+"/"+year, subreddit, schema, &processingProg)
		//scanDirForExtractedSubData(os.Getenv("BASE_DATA_DIRECTORY") + "/2016/Jan", "Basic")

	}
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()

	muxRouter.HandleFunc("/api", handleIndex).Methods("GET")
	muxRouter.HandleFunc("/api/status", handleStatus).Methods("GET", "OPTIONS")
	muxRouter.HandleFunc("/api/subs", handleGetSubs).Methods("GET")
	muxRouter.HandleFunc("/api/extractSub/{subreddit}", handleExtractSub).Methods("POST")
	muxRouter.HandleFunc("/api/status/{subreddit}", handleViewStatus).Methods("GET")
	muxRouter.HandleFunc("/api/processSub/{subreddit}", handleProcessSub).Methods("POST")
	muxRouter.HandleFunc("/api/mockStatus", handleMockStatus).Methods("GET")
	return muxRouter
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	writeStdHeaders(w)
	io.WriteString(w, "Welcome to the Reddit Comment Extractor!\nThese are the endpoints to use:\n")
	io.WriteString(w, "GET \"/api/status\" displays the overall status of this server and its data Processing\n")
	io.WriteString(w, "GET \"/api/subs\" lists all subreddits with all comments extracted and/or processed\n")
	io.WriteString(w, "POST \"/api/extractSub/<sub>\" extracts ALL comments from the <sub> subreddit, and saves to a datafile for later Processing\n")
	io.WriteString(w, "POST \"/api/processSub/<sub>\" processes the previously-extracted data for a subreddit, and saves these processed analytics for later retrieval\n")
	io.WriteString(w, "GET \"/api/status/<sub>\" displays the extraction and Processing status for a subreddit\n")
}

func handleMockStatus(w http.ResponseWriter, r *http.Request) {
	writeStdHeaders(w)
	status := ServerStatus{true, 65.67895, []string{"sample", "sample2"},
		true, 77.5678, []string{"sample3", "sample4"}}
	bytes, _ := json.Marshal(status)
	io.WriteString(w, string(bytes))
}

type ServerStatus struct {
	Processing      bool
	ProcessProgress float64
	ProcessQueue    []string

	Extracting      bool
	ExtractProgress float64
	ExtractQueue    []string
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	status := ServerStatus{}
	if len(extractSubQueue) > 0 {
		status.Extracting = true
		status.ExtractProgress = extractingProg
		status.ExtractQueue = extractSubQueue
	} else {
		status.Extracting = false
		status.ExtractProgress = 100
		status.ExtractQueue = make([]string, 0)
	}

	if len(processSubQueue) > 0 {
		status.Processing = true
		status.ProcessProgress = processingProg
		status.ProcessQueue = processSubQueue
	} else {
		status.Processing = false
		status.ProcessProgress = 100
		status.ProcessQueue = make([]string, 0)
	}
	bytes, _ := json.Marshal(status)

	writeStdHeaders(w)
	io.WriteString(w, string(bytes))
}

func handleGetSubs(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.Marshal(subredditStatuses)
	if err != nil {
		fmt.Println(err)
	}

	//for sub, v := range subredditStatuses {
	//	if v.ExtractedSummary != "" {
	//		io.WriteString(w, sub+"\n")
	//	}
	//}
	//io.WriteString(w, "\nSubreddits already processed:\n")
	//for sub, v := range subredditStatuses {
	//	if v.ProcessedSummary != "" {
	//		io.WriteString(w, sub+"\n")
	//	}
	//}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT, GET, POST")
	w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	io.WriteString(w, string(bytes))
}

func handleExtractSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subreddit := vars["subreddit"]

	if val, ok := subredditStatuses[subreddit]; ok {
		if val.Extracting || len(val.ExtractedMonthCommentCounts) != 0 {
			io.WriteString(w, "Subreddit "+subreddit+" has been extracted, is Extracting, or is in the queue for extraction!\n")
			io.WriteString(w, val.ToString())
		}
	} else {
		extractSubQueue = append(extractSubQueue, subreddit)
		io.WriteString(w, subreddit+" appended to extraction queue!\nNew job queue:\n")
		str := ""
		for i, v := range extractSubQueue {
			str += strconv.Itoa(i+1) + ". " + v + "\n"
		}

		subredditStatuses[subreddit] = subredditStatus{}

		if len(extractSubQueue) == 1 {
			go extractQueue() //this is the main goroutine that will extract all the future jobs
		}
		io.WriteString(w, str)
	}
	writeStdHeaders(w)
}

var extractingProg float64
var processingProg float64

func extractQueue() {
	tempSub := ""
	for len(extractSubQueue) > 0 {
		tempSub = extractSubQueue[0]
		if v, ok := subredditStatuses[tempSub]; !ok {
			subredditStatuses[tempSub] = subredditStatus{Extracting: true,
				ExtractedMonthCommentCounts: make(map[string]int64, 0), Processing: false, ProcessedSummary: ""}
		} else {
			v.Extracting = true
			subredditStatuses[tempSub] = v
		}
		summary := selection.SaveCriteriaDataToFile("subreddit", tempSub, "2016",
			os.Getenv("BASE_DATA_DIRECTORY"), selection.BasicSchema, &extractingProg)

		v := subredditStatuses[tempSub]
		v.Extracting = false

		v.ExtractedMonthCommentCounts = summary
		subredditStatuses[tempSub] = v
		extractSubQueue = extractSubQueue[1:] //done
		fmt.Println("COMPLETED")
	}
}

func handleProcessSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subreddit := vars["subreddit"]

	if val, ok := subredditStatuses[subreddit]; ok {
		if val.Processing {
			io.WriteString(w, "Subreddit \""+subreddit+"\" is still being processed!")
		} else if val.ProcessedSummary != "" {
			io.WriteString(w, "Subreddit \""+subreddit+"\" has already been processed:\n")
		} else if val.Extracting {
			io.WriteString(w, "Subreddit \""+subreddit+"\" is still being extracted")
		} else if len(val.ExtractedMonthCommentCounts) != 0 {
			processSubQueue = append(processSubQueue, subreddit)

			io.WriteString(w, subreddit+" appended to Processing queue!\nNew job queue:\n")
			str := ""
			for i, v := range processSubQueue {
				str += strconv.Itoa(i+1) + ". " + v + "\n"
			}
			subredditStatuses[subreddit] = subredditStatus{}

			if len(processSubQueue) == 1 {
				go processQueue() //this is the main goroutine that will process all the future jobs
			}
			io.WriteString(w, str)
		} else {
			io.WriteString(w, "Subreddit has not been extracted or processed yet, please hit the endpoint \"/extractSub/"+subreddit+"\" to extract the subreddit data first")
		}
	} else {
		io.WriteString(w, "Subreddit has not been extracted or processed yet, please hit the endpoint \"/extractSub/"+subreddit+"\" to extract the subreddit data first")
	}
	writeStdHeaders(w)
}

func processQueue() {
	tempSub := ""
	for len(processSubQueue) > 0 {
		tempSub = processSubQueue[0]
		if v, ok := subredditStatuses[tempSub]; !ok {
			log.Fatal("Tried to process a sub not found in the registry!")
		} else {
			v.Processing = true
			subredditStatuses[tempSub] = v
		}
		str := selection.OpenExtractedSubredditDatafile(os.Getenv("BASE_DATA_DIRECTORY")+"/"+"2016", tempSub, "Basic", &processingProg)

		v := subredditStatuses[tempSub]
		v.Processing = false
		v.ProcessedSummary = str
		subredditStatuses[tempSub] = v
		processSubQueue = processSubQueue[1:] //done
		fmt.Println("COMPLETED")
	}
}

func handleViewStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subreddit := vars["subreddit"]

	if val, ok := subredditStatuses[subreddit]; !ok {
		io.WriteString(w, "{}")
	} else {
		data, _ := json.Marshal(val)
		io.WriteString(w, string(data))
	}
	writeStdHeaders(w)
}

func writeStdHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT, GET, POST")
	w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
