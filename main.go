package main

import (
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

type SubredditJob struct {
	Year      string
	Month     string
	Subreddit string
}

//TODO move this out of main!
var extractSubQueue = make([]SubredditJob, 0)
var processSubQueue = make([]SubredditJob, 0)
var subredditStatuses = make(map[string]subredditStatus)

type subredditStatus struct {
	Extracting                  bool
	Extracted                   bool
	ExtractedMonthCommentCounts map[string]map[string]int64
	Processing                  bool
	Processed                   bool
	ProcessedSummary            selection.ProcessedSubredditStats
}

var DataDirectory string
var RunServer string
var ServerPort string
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
		RunServer = os.Getenv("RUN_SERVER")
		ServerPort = os.Getenv("SERVER_PORT")
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
		checkForExtractedSubs(yr, "Basic")
	}

	for _, subredditStatus := range subredditStatuses {
		for yearIdx, monthsAvailableArray := range YearsAndMonthsAvailable {
			//Ensure all the available years are in the status
			if _, ok := subredditStatus.ExtractedMonthCommentCounts[yearIdx]; !ok {
				subredditStatus.ExtractedMonthCommentCounts[yearIdx] = make(map[string]int64, 0)
			}

			for _, month := range selection.AllMonths {
				found := false
				for _, monAvailable := range monthsAvailableArray {
					if monthToShortIntString(monAvailable) == month {
						found = true
						break
					}
				}

				//the month isn't available
				if !found {
					subredditStatus.ExtractedMonthCommentCounts[yearIdx][month] = -1
					//fmt.Println("Marked " + yearIdx + "/" + month + " as unavailable")
				}

			}

		}

	}

	if RunServer == "true" {
		log.Fatal(run(ServerPort))
	} else {
		year := "2016"
		subreddit := "funny"
		schema := "Basic"
		//var prog float64
		//_ = selection.SaveCriteriaDataToFile("Subreddit", "funny", "2016", os.Getenv("BASE_DATA_DIRECTORY"), selection.BasicSchema, &prog)
		selection.OpenExtractedSubredditDatafile(DataDirectory, "Jan", year, subreddit, schema, &processingProg)
		//scanDirForExtractedSubData(os.Getenv("BASE_DATA_DIRECTORY") + "/2016/Jan", "Basic")

	}
}
func monthToShortIntString(month string) string {
	switch month {
	case "01":
		return "Jan"
	case "02":
		return "Feb"
	case "03":
		return "Mar"
	case "04":
		return "Apr"
	case "05":
		return "May"
	case "06":
		return "Jun"
	case "07":
		return "Jul"
	case "08":
		return "Aug"
	case "09":
		return "Sep"
	case "10":
		return "Oct"
	case "11":
		return "Nov"
	case "12":
		return "Dec"
	default:
		log.Println("Invalid month supplied; defaulting to january")
		return "01"
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

	go handleMessages()

	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func checkForExtractedSubs(year string, schema string) {

	yearDirectory := DataDirectory + "/Extracted/" + year + "/"

	for _, month := range selection.AllMonths {
		arr := selection.ScanDirForExtractedSubData(yearDirectory+month, schema) //scan the current Month folder for all subs
		if len(arr) == 0 {

		}

		for _, sub := range arr {
			fmt.Println("Checking " + sub)
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
				if val.ExtractedMonthCommentCounts[year] == nil {
					val.ExtractedMonthCommentCounts[year] = make(map[string]int64, 1)
				}
				val.ExtractedMonthCommentCounts[year][month] = commentCount
			} else {
				status := subredditStatus{Extracting: false, ExtractedMonthCommentCounts: make(map[string]map[string]int64, 1), Processing: false, ProcessedSummary: selection.ProcessedSubredditStats{}}
				status.ExtractedMonthCommentCounts[year] = make(map[string]int64, 1)
				status.ExtractedMonthCommentCounts[year][month] = commentCount
				subredditStatuses[sub] = status
			}
		}
	}
}

func (status subredditStatus) ToString() string {
	str := "Extracting: " + strconv.FormatBool(status.Extracting) + "\n"

	if !status.Extracting && len(status.ExtractedMonthCommentCounts) != 0 {
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
	muxRouter.HandleFunc("/api/extractSub/{Subreddit}/{Month}/{Year}", handleExtractSub).Methods("POST")
	muxRouter.HandleFunc("/api/status/{Subreddit}", handleViewStatus).Methods("GET")
	muxRouter.HandleFunc("/api/processSub/{Subreddit}/{Month}/{Year}", handleProcessSub).Methods("POST")
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
	ProcessQueue    []SubredditJob

	Extracting      bool
	ExtractProgress float64
	ExtractQueue    []SubredditJob
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
		status.ExtractQueue = make([]SubredditJob, 0)
	}

	if len(processSubQueue) > 0 {
		status.Processing = true
		status.ProcessProgress = processingProg
		status.ProcessQueue = processSubQueue
	} else {
		status.Processing = false
		status.ProcessProgress = 100
		status.ProcessQueue = make([]SubredditJob, 0)
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
	subreddit := vars["Subreddit"]
	month := vars["Month"]
	year := vars["Year"]
	writeStdHeaders(w)
	//if _, ok := subredditStatuses[Subreddit]; ok {
	//	// TODO fix this
	//	//if val.Extracting || len(val.ExtractedMonthCommentCounts) != 0 {
	//	//	io.WriteString(w, "Subreddit "+Subreddit+" has been extracted, is Extracting, or is in the queue for extraction!\n")
	//	//	io.WriteString(w, val.ToString())
	//	//}
	//} else {
	job := SubredditJob{Month: month, Year: year, Subreddit: subreddit}
	extractSubQueue = append(extractSubQueue, job)
	io.WriteString(w, subreddit+" in "+month+"/"+year+" appended to extraction queue!\nNew job queue:\n")
	str := ""
	for i, v := range extractSubQueue {
		str += strconv.Itoa(i+1) + ". " + v.Subreddit + "\n"
	}

	//subredditStatuses[subreddit] = subredditStatus{}

	if len(extractSubQueue) == 1 {
		go extractQueue() //this is the main goroutine that will extract all the future jobs
	}
	io.WriteString(w, str)
	//}
}

var extractingProg float64
var processingProg float64

func extractQueue() {
	var tempSub SubredditJob
	for len(extractSubQueue) > 0 {
		tempSub = extractSubQueue[0]
		if v, ok := subredditStatuses[tempSub.Subreddit]; !ok {
			subredditStatuses[tempSub.Subreddit] = subredditStatus{Extracting: true,
				ExtractedMonthCommentCounts: make(map[string]map[string]int64, 0), Processing: false, ProcessedSummary: selection.ProcessedSubredditStats{}}
		} else {
			v.Extracting = true
			subredditStatuses[tempSub.Subreddit] = v
			extractingProg = 0
		}
		go monitorProgress(&extractingProg)

		//TODO ensure that the Month/Year for extraction is present in the list of uncompressed data entries
		summary := selection.ExtractCriteriaDataToFile("Subreddit", tempSub.Subreddit, tempSub.Year, tempSub.Month,
			DataDirectory, selection.BasicSchema, &extractingProg)

		v := subredditStatuses[tempSub.Subreddit]
		v.Extracting = false
		v.Extracted = true
		if v.ExtractedMonthCommentCounts == nil {
			v.ExtractedMonthCommentCounts = make(map[string]map[string]int64, 1)
		}
		if v.ExtractedMonthCommentCounts[tempSub.Year] == nil {
			v.ExtractedMonthCommentCounts[tempSub.Year] = make(map[string]int64, 1)
		}
		v.ExtractedMonthCommentCounts[tempSub.Year][tempSub.Month] = summary
		subredditStatuses[tempSub.Subreddit] = v
		extractSubQueue = extractSubQueue[1:] //done
		fmt.Println("COMPLETED")
		broadcast <- "fetch"
	}
}

func handleProcessSub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subreddit := vars["Subreddit"]
	month := vars["Month"]
	year := vars["Year"]

	writeStdHeaders(w)
	if val, ok := subredditStatuses[subreddit]; ok {
		if val.Processing {
			io.WriteString(w, "Subreddit \""+subreddit+"\" is still being processed!")
		} else if val.Processed {
			io.WriteString(w, "Subreddit \""+subreddit+"\" has already been processed:\n")
		} else if val.Extracting {
			io.WriteString(w, "Subreddit \""+subreddit+"\" is still being extracted")
		} else {
			job := SubredditJob{Month: month, Year: year, Subreddit: subreddit}
			processSubQueue = append(processSubQueue, job)

			io.WriteString(w, "{ "+subreddit+" appended to Processing queue!\nNew job queue:\n")
			str := ""
			for i, v := range processSubQueue {
				str += strconv.Itoa(i+1) + ". " + v.Subreddit + "\n"
			}

			if len(processSubQueue) == 1 {
				go processQueue() //this is the main goroutine that will process all the future jobs
			}
			io.WriteString(w, str+"}")
		}
	} else {
		io.WriteString(w, "Subreddit has not been extracted or processed yet, please hit the endpoint \"/extractSub/"+subreddit+"\" to extract the Subreddit data first")
	}
}

func processQueue() {
	var tempSub SubredditJob
	for len(processSubQueue) > 0 {
		tempSub = processSubQueue[0]
		if v, ok := subredditStatuses[tempSub.Subreddit]; !ok {
			log.Fatal("Tried to process a sub not found in the registry!")
		} else {
			v.Processing = true
			processingProg = 0
			subredditStatuses[tempSub.Subreddit] = v
		}

		go monitorProgress(&processingProg)
		sum := selection.OpenExtractedSubredditDatafile(DataDirectory, tempSub.Month, tempSub.Year, tempSub.Subreddit, "Basic", &processingProg)

		v := subredditStatuses[tempSub.Subreddit]
		v.Processing = false
		v.Processed = true
		v.ProcessedSummary = sum
		subredditStatuses[tempSub.Subreddit] = v
		processSubQueue = processSubQueue[1:] //done
		fmt.Println("COMPLETED")
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
		subredditStatuses[subreddit] = subredditStatus{Extracting: false, ExtractedMonthCommentCounts: make(map[string]map[string]int64, 0), Processing: false, ProcessedSummary: selection.ProcessedSubredditStats{}}
		for yearIdx, monthsAvailableArray := range YearsAndMonthsAvailable {
			//Ensure all the available years are in the status
			if _, ok := subredditStatuses[subreddit].ExtractedMonthCommentCounts[yearIdx]; !ok {
				subredditStatuses[subreddit].ExtractedMonthCommentCounts[yearIdx] = make(map[string]int64, 0)
			}

			for _, month := range selection.AllMonths {
				found := false
				for _, monAvailable := range monthsAvailableArray {
					if monthToShortIntString(monAvailable) == month {
						found = true
						break
					}
				}

				//the month isn't available
				if !found {
					subredditStatuses[subreddit].ExtractedMonthCommentCounts[yearIdx][month] = -1
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

func writeStdHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT, GET, POST")
	w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
