package selection

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// ListYearsAndMonthsForExtractionInDir scans the directory provided, and returns a map that represents the years
// and months of data available in the directory that are ready for extraction (raw, uncompressed data files of
// the format "RC_2016-05" for May of 2016, for example)
func ListYearsAndMonthsForExtractionInDir(dir string) map[string][]string {
	yearsAndMonths := make(map[string][]string, 0)

	f, err := os.Open(dir)
	if err != nil {
		log.Fatal(err)
	}
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			if strings.HasPrefix(file.Name(), "RC_") && !strings.Contains(file.Name(), ".bz2") &&
				!strings.Contains(file.Name(), "count") && !strings.Contains(file.Name(), ".txt") {

				yrMonth := strings.Replace(file.Name(), "RC_", "", -1)
				fields := strings.Split(yrMonth, "-")
				if yearsAndMonths[fields[0]] == nil {
					yearsAndMonths[fields[0]] = make([]string, 0)
				}
				yearsAndMonths[fields[0]] = append(yearsAndMonths[fields[0]], fields[1])
			}
		}
	}
	return yearsAndMonths
}

func readInCommentCountMetadata(rawDataFilePath string) int64 {
	linecountFilePath := rawDataFilePath + "_count"
	if _, err := os.Stat(linecountFilePath); !os.IsNotExist(err) {
		var count int64
		plan, fileOpenErr := ioutil.ReadFile(linecountFilePath)
		if fileOpenErr != nil {
			log.Fatal("failed to open " + linecountFilePath)
		}
		err := json.Unmarshal(plan, &count)
		if err != nil {
			log.Fatal(err)
		}
		return count
	}
	return 0
}

func monthToIntString(month string) string {
	switch month {
	case "Jan":
		return "-01"
	case "Feb":
		return "-02"
	case "Mar":
		return "-03"
	case "Apr":
		return "-04"
	case "May":
		return "-05"
	case "Jun":
		return "-06"
	case "Jul":
		return "-07"
	case "Aug":
		return "-08"
	case "Sep":
		return "-09"
	case "Oct":
		return "-10"
	case "Nov":
		return "-11"
	case "Dec":
		return "-12"
	default:
		log.Println("Invalid month supplied; defaulting to january")
		return "-01"
	}
}
