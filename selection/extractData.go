package selection

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/valyala/fastjson"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

//TODO explore speedups related to unravelling the recursion
func recurseBuildCompleteLine(reader *bufio.Reader) []byte {
	line, isPrefix, err := reader.ReadLine()
	if err != nil {
		log.Println(err)
		return nil
	}

	if isPrefix {
		return append(line, recurseBuildCompleteLine(reader)...)
	} else {
		return line
	}
}

func getCommentDataFromLine(line []byte, keyTypes map[string]string) map[string]string {
	result := make(map[string]string, 0)

	for key, v := range keyTypes {
		if v == "str" {
			result[key] = strings.ToLower(fastjson.GetString(line, key))
		} else if v == "int" {
			result[key] = strconv.Itoa(fastjson.GetInt(line, key))
		} else {
			log.Fatal("Undetected type: " + v)
		}
	}
	return result
}

const percentPerMonth = (1.0 / 12.0) * 100.0

func SaveCriteriaDataToFile(criteria string, value string, year string, basedir string, schema commentSchema, progress *float64) string {
	summary := ""
	criteria = strings.ToLower(criteria)
	value = strings.ToLower(value)
	for mIndex, v := range AllMonths {
		relevantComments := make([]map[string]string, 0)
		str := basedir + "/" + year + "/" + v + "/" + criteria + "_" + value + "_" + schema.name
		if _, err := os.Stat(str); !os.IsNotExist(err) {
			fmt.Println("Found cached data for " + str)
			summary += "Found cached data for " + str + "\n"
			continue
		} else {
			fmt.Println("No cached data found for " + str)
		}

		var buffer bytes.Buffer
		buffer.Write([]byte(basedir + "/" + year + "/RC_" + year + monthToIntString(v)))
		file, fileOpenErr := os.Open(buffer.String())
		if fileOpenErr != nil {
			fmt.Print(fileOpenErr)
			os.Exit(0)
		} else {
			fmt.Println("Opened file " + buffer.String())
		}

		var totalLines uint64
		metafile, fileOpenErr2 := os.Open(buffer.String() + "_meta.txt")
		if fileOpenErr2 != nil {
			totalLines = 0
			metafile.Close()
		} else {
			rdr := bufio.NewReader(metafile)
			line, _, _ := rdr.ReadLine()
			totalLines, _ = strconv.ParseUint(string(line), 10, 64)
		}

		reader := bufio.NewReaderSize(file, 4096)

		var linesRead uint64 = 0
		startTime := time.Now()
		tempTime := startTime
		for {
			line := recurseBuildCompleteLine(reader)
			if line == nil {
				fmt.Println("Lines: " + strconv.FormatUint(linesRead, 10))
				log.Println("Encountered error; concluding analysis")
				break
			}
			linesRead++

			if strings.Contains(strings.ToLower(string(line)), "\""+criteria+"\":\""+value+"\"") {
				parsed := getCommentDataFromLine(line, schema.schema)
				relevantComments = append(relevantComments, parsed)
			}

			if linesRead%HundredThousand == 0 {
				if totalLines != 0 {
					percentDone := (float64(linesRead) / float64(totalLines)) * 100.0
					fmt.Println("Percent complete with " + v + ": " + strconv.FormatFloat(percentDone, 'f', 2, 64) + "%")
					*progress = (percentPerMonth * float64(mIndex)) + (percentPerMonth * (percentDone / 100.0))
				} else {
					fmt.Println("Lines complete: " + strconv.FormatUint(linesRead, 10))
				}
				fmt.Println("Processed 100k lines in " + time.Now().Sub(tempTime).String())
				tempTime = time.Now()
			}
		}
		dif := time.Now().Sub(startTime).String()
		tempStr := "Took " + dif + " to search " + strconv.FormatUint(linesRead, 10) + " comments of file " + buffer.String() + "\n"
		if len(relevantComments) == 0 {
			log.Println("Found 0 comments for " + criteria + ":" + value + " in " + v + ", exiting extraction!")
			return "ERROR"
		}
		summary += tempStr
		fmt.Println(tempStr)
		dumpDataToFilepath(relevantComments, str)
		*progress = percentPerMonth * float64(mIndex+1.0)
	}
	return summary
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
