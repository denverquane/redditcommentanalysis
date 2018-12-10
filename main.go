package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/denverquane/redditcommentanalysis/selection"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan string)            // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SubredditProcessJob struct {
	Year       string
	Month      string
	Subreddits []string
}

type SubredditExtractJob struct {
	Year       string
	Month      string
	Subreddits []string
}

//TODO move this out of main!
var extractSubQueue = make([]SubredditExtractJob, 0)
var processSubQueue = make([]SubredditProcessJob, 0)
var subredditStatuses = make(map[string]subredditStatus)

type subredditStatus struct {
	Extracting                      bool
	Extracted                       bool
	ExtractedYearMonthCommentCounts map[string]map[string]int64

	Processing                         bool
	Processed                          bool
	ProcessedYearMonthCommentSummaries map[string]map[string]selection.ProcessedSubredditStats
}

var DataDirectory string
var ServerPort string
var CertDirectory string
var YearsAndMonthsAvailable map[string][]string

func main() {

	err := godotenv.Load()

	// TODO if the env can't be loaded, check the REDDIT_DATA_DIRECTORY env var instead

	// TODO if the REDDIT_DATA_DIRECTORY doesn't work, assume in a docker container, and use /data

	// TODO if /data doesn't work (or have the right format?) exit the program with error

	if err != nil {
		log.Fatal("Error loading .env file")
	} else {
		DataDirectory = os.Getenv("BASE_DATA_DIRECTORY")
		ServerPort = os.Getenv("SERVER_PORT")
		CertDirectory = os.Getenv("CERT_DIRECTORY")
		log.Println("Cert directory: " + CertDirectory)
	}

	YearsAndMonthsAvailable = selection.ListYearsAndMonthsForExtractionInDir(DataDirectory)
	for i, v := range YearsAndMonthsAvailable {
		fmt.Print(i + " has [ ")
		for _, vv := range v {
			fmt.Print(vv + " ")
		}
		fmt.Println("] to extract ")
	}

	for yr := range YearsAndMonthsAvailable {
		//fmt.Println("Checking " + yr)
		checkForExtractedSubs(yr, "Best")
		checkForProcessedData(yr)
	}

	for _, subredditStatus := range subredditStatuses {
		for yearIdx, monthsAvailableArray := range YearsAndMonthsAvailable {
			//Ensure all the available years are in the status
			if _, ok := subredditStatus.ExtractedYearMonthCommentCounts[yearIdx]; !ok {
				subredditStatus.ExtractedYearMonthCommentCounts[yearIdx] = make(map[string]int64, 0)
			}

			for _, month := range selection.AllMonths {
				found := false
				for _, monAvailable := range monthsAvailableArray {
					if selection.MonthToShortIntString(monAvailable) == month {
						found = true
						break
					}
				}

				//the month isn't available
				if !found {
					subredditStatus.ExtractedYearMonthCommentCounts[yearIdx][month] = -1
					//fmt.Println("Marked " + yearIdx + "/" + month + " as unavailable")
				}
			}
		}
	}

	log.Fatal(run(ServerPort))
}

func run(port string) error {
	cer, err := tls.LoadX509KeyPair(CertDirectory+"/fullchain.pem", CertDirectory+"/privkey.pem")
	if err != nil {
		log.Println(err)
		return err
	}

	handler := makeMuxRouter()

	s := &http.Server{
		Addr:           ":https",
		Handler:        handler,
		TLSConfig:      &tls.Config{Certificates: []tls.Certificate{cer}},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go handleMessages()

	go http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
	}))
	log.Fatal(s.ListenAndServeTLS("", "")) //Key and cert are already set in the TLS config
	return nil
}

func checkForProcessedData(year string) {
	yearDirectory := DataDirectory + "/Processed/" + year + "/"

	for _, month := range selection.AllMonths {
		arr := selection.ScanDirForProcessedSubData(yearDirectory + month) //scan the current Month folder for all subs
		log.Println("Checking " + year + "/" + month + " for processed data")
		for _, sub := range arr {

			var subredditSummary selection.ProcessedSubredditStats
			str := yearDirectory + month + "/" + "subreddit_" + sub

			f, err := os.Open(str)
			if err != nil {
				log.Println(err)
			}

			bytess, err := ioutil.ReadAll(f)
			if err == nil {
				err2 := json.Unmarshal(bytess, &subredditSummary)
				if err2 != nil {
					panic(err2)
				}
			} else {
				panic(err)
			}

			f.Close()

			if val, ok := subredditStatuses[sub]; ok {
				if val.ProcessedYearMonthCommentSummaries[year] == nil {
					val.ProcessedYearMonthCommentSummaries[year] = make(map[string]selection.ProcessedSubredditStats, 1)
				}
				val.ProcessedYearMonthCommentSummaries[year][month] = subredditSummary
				val.Processed = true
				subredditStatuses[sub] = val
			} else {
				status := subredditStatus{Extracting: false, Extracted: false, ExtractedYearMonthCommentCounts: make(map[string]map[string]int64, 0), Processing: false, Processed: true, ProcessedYearMonthCommentSummaries: make(map[string]map[string]selection.ProcessedSubredditStats, 1)}
				status.ProcessedYearMonthCommentSummaries[year] = make(map[string]selection.ProcessedSubredditStats, 1)
				status.ProcessedYearMonthCommentSummaries[year][month] = subredditSummary
				subredditStatuses[sub] = status
			}
		}
	}
}

func checkForExtractedSubs(year string, schema string) {

	yearDirectory := DataDirectory + "/Extracted/" + year + "/"

	for _, month := range selection.AllMonths {
		arr := selection.ScanDirForExtractedSubData(yearDirectory+month, schema) //scan the current Month folder for all subs
		log.Println("Checking " + year + "/" + month + " for extracted data")
		for _, sub := range arr {
			var commentCount int64
			str := yearDirectory + month + "/subreddit_" + sub + "_" + schema + "_count"
			plan, fileOpenErr := ioutil.ReadFile(str)
			if fileOpenErr != nil {
				log.Println("Failed to open " + str + ", now opening datafile and writing total comment count to file")
				var comments []map[string]string
				str := yearDirectory + month + "/" + "subreddit_" + sub + "_" + schema

				f, err := os.Open(str)
				if err != nil {
					log.Println(err)
				}
				lines, err2 := selection.LineCounter(f)
				if err2 != nil {
					log.Println(err2)
				}

				file, err2 := os.Create(yearDirectory + month + "/" + "subreddit_" + sub + "_" + schema + "_count")
				if err != nil {
					log.Println(err2)
				}
				length := strconv.FormatInt(int64(lines), 10)
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
				if val.ExtractedYearMonthCommentCounts[year] == nil {
					val.ExtractedYearMonthCommentCounts[year] = make(map[string]int64, 1)
				}
				val.ExtractedYearMonthCommentCounts[year][month] = commentCount
				subredditStatuses[sub] = val
			} else {
				status := subredditStatus{Extracting: false, Extracted: true, ExtractedYearMonthCommentCounts: make(map[string]map[string]int64, 1), Processing: false, ProcessedYearMonthCommentSummaries: make(map[string]map[string]selection.ProcessedSubredditStats, 0)}
				status.ExtractedYearMonthCommentCounts[year] = make(map[string]int64, 1)
				status.ExtractedYearMonthCommentCounts[year][month] = commentCount
				subredditStatuses[sub] = status
			}
		}
	}
}

func (status subredditStatus) ToString() string {
	str := "Extracting: " + strconv.FormatBool(status.Extracting) + "\n"

	if !status.Extracting && len(status.ExtractedYearMonthCommentCounts) != 0 {
		str += "Extracted summary: " + "\n"
	}

	str += "Processing: " + strconv.FormatBool(status.Processing) + "\n"

	//if !status.Processing && status.ProcessedSummary != "" {
	//	str += "Processed summary: " + status.ProcessedSummary + "\n"
	//}

	return str
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()

	muxRouter.HandleFunc("/api", handleIndex).Methods("GET")
	muxRouter.HandleFunc("/api/status", handleStatus).Methods("GET", "OPTIONS")
	muxRouter.HandleFunc("/api/subs", handleGetSubs).Methods("GET")
	muxRouter.HandleFunc("/api/extractSubs/{Month}/{Year}", handleExtractSubs).Methods("POST")
	muxRouter.HandleFunc("/api/status/{Subreddit}", handleViewStatus).Methods("GET")
	muxRouter.HandleFunc("/api/processSub/{Subreddit}/{Month}/{Year}", handleProcessSub).Methods("POST")
	//muxRouter.HandleFunc("/api/combineProcessed/{Year}", combineProcessed).Methods("POST")
	muxRouter.HandleFunc("/api/addSubEntry/{Subreddit}", addSubredditEntry).Methods("POST")
	muxRouter.HandleFunc("/api/mockStatus", handleMockStatus).Methods("GET")
	muxRouter.HandleFunc("/ws", handleConnections)
	return muxRouter
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	writeStdHeaders(w)
	io.WriteString(w, "Welcome to the Reddit Comment Extractor!\nThese are the endpoints to use:\n")
	io.WriteString(w, "GET \"/api/status\" displays the overall status of this server and its data Processing\n")
	io.WriteString(w, "GET \"/api/subs\" lists all subreddits with all comments extracted and/or processed\n")
	io.WriteString(w, "POST \"/api/extractSub/<sub>/<Month>/<Year>\" extracts ALL comments from the <sub> Subreddit, and saves to a datafile for later Processing\n")
	io.WriteString(w, "POST \"/api/processSub/<sub>\" processes the previously-extracted data for a Subreddit, and saves these processed analytics for later retrieval\n")
	io.WriteString(w, "GET \"/api/status/<sub>\" displays the extraction and Processing status for a Subreddit\n")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	for {
		var msg bool
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		//fmt.Println("Have msg to transmit")
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func handleMockStatus(w http.ResponseWriter, r *http.Request) {
	//writeStdHeaders(w)
	//status := ServerStatus{true, 65.67895, []string{"sample", "sample2"},
	//	true, 77.5678, []string{"sample3", "sample4"}}
	//bytes, _ := json.Marshal(status)
	//io.WriteString(w, string(bytes))
}

type ServerStatus struct {
	Processing      bool
	ProcessProgress float64
	ProcessQueue    []SubredditProcessJob

	Extracting      bool
	ExtractProgress float64
	ExtractQueue    []SubredditExtractJob
	ExtractTimeRem  string
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	status := ServerStatus{}
	if len(extractSubQueue) > 0 {
		status.Extracting = true
		status.ExtractProgress = extractingProg
		status.ExtractQueue = extractSubQueue
		status.ExtractTimeRem = extractTimeRemaining
	} else {
		status.Extracting = false
		status.ExtractProgress = 100
		status.ExtractQueue = make([]SubredditExtractJob, 0)
		status.ExtractTimeRem = ""
	}

	if len(processSubQueue) > 0 {
		status.Processing = true
		status.ProcessProgress = processingProg
		status.ProcessQueue = processSubQueue
	} else {
		status.Processing = false
		status.ProcessProgress = 100
		status.ProcessQueue = make([]SubredditProcessJob, 0)
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

func handleExtractSubs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	month := vars["Month"]
	year := vars["Year"]
	writeStdHeaders(w)

	decoder := json.NewDecoder(r.Body)

	var data []string
	err := decoder.Decode(&data)
	if err != nil {
		fmt.Println("Error dsec")
		panic(err)
	}

	job := SubredditExtractJob{Month: month, Year: year, Subreddits: data}
	extractSubQueue = append(extractSubQueue, job)
	//io.WriteString(w, data+" in "+month+"/"+year+" appended to extraction queue!\nNew job queue:\n")
	str := ""
	for i, v := range extractSubQueue {
		str += strconv.Itoa(i+1) + ". " + v.Year + "/" + v.Month + "\n"
	}

	if len(extractSubQueue) == 1 {
		go extractQueue() //this is the main goroutine that will extract all the future jobs
	}
	io.WriteString(w, str)
}

var extractingProg float64
var extractTimeRemaining string
var processingProg float64

func extractQueue() {
	var tempSub SubredditExtractJob
	for len(extractSubQueue) > 0 {
		tempSub = extractSubQueue[0]

		for _, sub := range tempSub.Subreddits {
			if v, ok := subredditStatuses[sub]; !ok {
				subredditStatuses[sub] = subredditStatus{Extracting: true,
					ExtractedYearMonthCommentCounts: make(map[string]map[string]int64, 0), Processing: false, ProcessedYearMonthCommentSummaries: make(map[string]map[string]selection.ProcessedSubredditStats, 0)}
			} else {
				v.Extracting = true
				subredditStatuses[sub] = v
				extractingProg = 0
			}
		}

		go monitorProgress(&extractingProg)

		//TODO ensure that the Month/Year for extraction is present in the list of uncompressed data entries

		criterias := make([]selection.Criteria, len(tempSub.Subreddits))

		for i, v := range tempSub.Subreddits {
			criterias[i].Value = v
			criterias[i].Test = "subreddit"
		}
		summary := selection.ExtractCriteriaDataToFile(criterias, tempSub.Year, tempSub.Month,
			DataDirectory, selection.BestSchema, &extractingProg, &extractTimeRemaining)

		for i, sub := range tempSub.Subreddits {
			v := subredditStatuses[sub]
			v.Extracting = false
			v.Extracted = true
			if v.ExtractedYearMonthCommentCounts == nil {
				v.ExtractedYearMonthCommentCounts = make(map[string]map[string]int64, 1)
			}
			if v.ExtractedYearMonthCommentCounts[tempSub.Year] == nil {
				v.ExtractedYearMonthCommentCounts[tempSub.Year] = make(map[string]int64, 1)
			}
			v.ExtractedYearMonthCommentCounts[tempSub.Year][tempSub.Month] = int64(summary[i])
			subredditStatuses[sub] = v
		}

		extractSubQueue = extractSubQueue[1:] //done
		fmt.Println("COMPLETED")
		checkForExtractedSubs(tempSub.Year, "Best")
		broadcast <- "fetch"
	}
}

func handleProcessSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subreddit := vars["Subreddit"]
	month := vars["Month"]
	year := vars["Year"]
	override := false

	writeStdHeaders(w)
	if val, ok := subredditStatuses[subreddit]; ok {
		if val.Extracting {
			io.WriteString(w, "Subreddit \""+subreddit+"\" is still being extracted")
		} else {
			arr := make([]string, 1)
			arr[0] = subreddit

			if month == "ALL" {
				if len(processSubQueue) == 0 {
					override = true
				}
				for _, v := range selection.AllMonths {
					job := SubredditProcessJob{Month: v, Year: year, Subreddits: arr}
					processSubQueue = append(processSubQueue, job)
				}
			} else {
				job := SubredditProcessJob{Month: month, Year: year, Subreddits: arr}
				processSubQueue = append(processSubQueue, job)
			}

			io.WriteString(w, "{ "+subreddit+" appended to Processing queue!\nNew job queue:\n")
			str := ""
			for i, v := range processSubQueue {
				str += strconv.Itoa(i+1) + ". " + v.Month + "/" + v.Year + "\n"
			}

			if len(processSubQueue) == 1 || override {
				go processQueue() //this is the main goroutine that will process all the future jobs
			}
			io.WriteString(w, str+"}")
		}
	} else {
		io.WriteString(w, "Subreddit has not been extracted or processed yet, please hit the endpoint \"/extractSub/"+subreddit+"\" to extract the Subreddit data first")
	}
}

func processQueue() {
	var tempSub SubredditProcessJob
	log.Println("Started new process worker!")
	for len(processSubQueue) > 0 {
		tempSub = processSubQueue[0]

		go monitorProgress(&processingProg)

		for _, sub := range tempSub.Subreddits {
			if subredditStatuses[sub].ProcessedYearMonthCommentSummaries != nil &&
				subredditStatuses[sub].ProcessedYearMonthCommentSummaries[tempSub.Year] != nil {
				if _, ok := subredditStatuses[sub].ProcessedYearMonthCommentSummaries[tempSub.Year][tempSub.Month]; ok {
					log.Println("Subreddit " + sub + " has already been processed! Skipping!")
					continue
				}
			}

			sum := selection.OpenExtractedSubredditDatafile(DataDirectory, tempSub.Month, tempSub.Year, sub, "Best", &processingProg)
			v := subredditStatuses[sub]
			v.Processing = false
			v.Processed = true
			if v.ProcessedYearMonthCommentSummaries == nil {
				v.ProcessedYearMonthCommentSummaries = make(map[string]map[string]selection.ProcessedSubredditStats, 1)
			}
			if v.ProcessedYearMonthCommentSummaries[tempSub.Year] == nil {
				v.ProcessedYearMonthCommentSummaries[tempSub.Year] = make(map[string]selection.ProcessedSubredditStats, 1)
			}
			v.ProcessedYearMonthCommentSummaries[tempSub.Year][tempSub.Month] = sum
			subredditStatuses[sub] = v
		}

		processSubQueue = processSubQueue[1:] //done
		fmt.Println("COMPLETED")
		checkForProcessedData(tempSub.Year)
		broadcast <- "fetch"
	}
}

func monitorProgress(prog *float64) {
	prev := 0.1
	for {
		if *prog == 100.0 {
			break
		} else if (*prog) != prev {
			prev = *prog
			broadcast <- "status"
		}
		time.Sleep(time.Second)
	}
}

func addSubredditEntry(w http.ResponseWriter, r *http.Request) {
	writeStdHeaders(w)
	vars := mux.Vars(r)
	subreddit := vars["Subreddit"]

	if _, ok := subredditStatuses[subreddit]; !ok {
		subredditStatuses[subreddit] = subredditStatus{Extracting: false, ExtractedYearMonthCommentCounts: make(map[string]map[string]int64, 0), Processing: false, ProcessedYearMonthCommentSummaries: make(map[string]map[string]selection.ProcessedSubredditStats, 0)}
		for yearIdx, monthsAvailableArray := range YearsAndMonthsAvailable {
			//Ensure all the available years are in the status
			if _, ok := subredditStatuses[subreddit].ExtractedYearMonthCommentCounts[yearIdx]; !ok {
				subredditStatuses[subreddit].ExtractedYearMonthCommentCounts[yearIdx] = make(map[string]int64, 0)
			}

			for _, month := range selection.AllMonths {
				found := false
				for _, monAvailable := range monthsAvailableArray {
					if selection.MonthToShortIntString(monAvailable) == month {
						found = true
						break
					}
				}

				//the month isn't available
				if !found {
					subredditStatuses[subreddit].ExtractedYearMonthCommentCounts[yearIdx][month] = -1
					//fmt.Println("Marked " + yearIdx + "/" + month + " as unavailable")
				}

			}

		}
		io.WriteString(w, "{}")
	}

}

func handleViewStatus(w http.ResponseWriter, r *http.Request) {
	writeStdHeaders(w)
	vars := mux.Vars(r)
	subreddit := vars["Subreddit"]

	if val, ok := subredditStatuses[subreddit]; !ok {
		io.WriteString(w, "{}")
	} else {
		data, _ := json.Marshal(val)
		io.WriteString(w, string(data))
	}
}

//func combineProcessed(w http.ResponseWriter, r *http.Request) {
//	writeStdHeaders(w)
//	vars := mux.Vars(r)
//	year := vars["Year"]
//
//	selection.CombineAllToSingleCSV(DataDirectory, year, "Basic")
//}

func writeStdHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT, GET, POST")
	w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
