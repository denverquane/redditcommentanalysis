package selection

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/valyala/fastjson"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const HundredThousand = 100000
const OneMillion = 1000000
const TenMillion = 10000000

type SearchParams struct {
	year             string
	months           []string
	lineLimit        uint64
	stringORQueries  []string
	stringANDQueries []string
	intORQueries     []IntCriteria
	intANDQueries    []IntCriteria
}

func (params SearchParams) ToString() string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	buffer.WriteString("\nyear: " + params.year)
	buffer.WriteString("\nmonths: \n[")
	for _, v := range params.months {
		buffer.WriteString("\n" + v)
	}
	buffer.WriteString("\n]")
	buffer.WriteString("\nlinelimit: " + strconv.FormatUint(params.lineLimit, 10))
	buffer.WriteString("\nORQueries: \n[")
	for _, v := range params.stringORQueries {
		buffer.WriteString("\n" + v)
	}
	buffer.WriteString("\n]")
	buffer.WriteString("\nANDQueries: \n[")
	for _, v := range params.stringANDQueries {
		buffer.WriteString("\n" + v)
	}
	buffer.WriteString("\n]")
	return buffer.String()
}

func (params SearchParams) ValuesToString() string {
	var buffer bytes.Buffer
	buffer.WriteString(params.year + "_")
	for _, v := range params.months {
		buffer.WriteString(v + "_")
	}
	buffer.WriteString(strconv.FormatUint(params.lineLimit, 10) + "_")
	for _, v := range params.stringORQueries {
		str := strings.Replace(v, "\"", "", -1) + "_"
		buffer.WriteString(strings.Replace(str, ":", "", -1) + "_")
	}

	for _, v := range params.stringANDQueries {
		str := strings.Replace(v, "\"", "", -1) + "_"
		buffer.WriteString(strings.Replace(str, ":", "", -1) + "_")
	}
	return buffer.String()
}

type SimpleSearchCriteria struct {
	Year             string
	Months           []string
	LineLimit        int
	StringORQueries  []string
	StringANDQueries []string
}

func InterpretFromHttp(r *http.Request) *SearchParams {
	params := MakeEmptySearchParams()
	var message SimpleSearchCriteria
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&message); err != nil {
		fmt.Println(err)
	}
	fmt.Println(message)
	params.year = message.Year
	params.months = message.Months
	params.lineLimit = uint64(message.LineLimit)
	params.stringORQueries = message.StringORQueries
	params.stringANDQueries = message.StringANDQueries
	fmt.Println(params)
	return params
}

func MakeEmptySearchParams() *SearchParams {
	return &SearchParams{"", make([]string, 0), 0,
		make([]string, 0), make([]string, 0),
		make([]IntCriteria, 0), make([]IntCriteria, 0)}
}

func MakeSimpleSearchParams(year string, months []string, limit uint64, ORs []string, ANDs []string) *SearchParams {
	return &SearchParams{year, months, limit,
		ORs, ANDs,
		make([]IntCriteria, 0), make([]IntCriteria, 0)}
}

func (param SearchParams) IsReady() (bool, string) {
	if param.year == "" {
		return false, "You haven't specified a search year"
	} else if len(param.months) == 0 && param.lineLimit == 0 {
		return false, "You haven't specified a search month, or a line limit number!"
	}
	return true, "Ready!"
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

func OpenAndSearchFile(params SearchParams, baseDataDirectory string, commentFields map[string]string) []map[string]string {
	relevantComments := make([]map[string]string, 0)
	for _, v := range params.months {
		var buffer bytes.Buffer
		buffer.Write([]byte(baseDataDirectory))
		buffer.Write([]byte("/RC_" + params.year + monthToIntString(v)))
		file, fileOpenErr := os.Open(buffer.String())
		metafile, fileOpenErr2 := os.Open(buffer.String() + "_meta.txt")

		if fileOpenErr != nil {
			fmt.Print(fileOpenErr)
			os.Exit(0)
		}
		var totalLines uint64

		if fileOpenErr2 != nil {
			totalLines = 0
			metafile.Close()
		} else {
			rdr := bufio.NewReader(metafile)
			line, _, _ := rdr.ReadLine()
			totalLines, _ = strconv.ParseUint(string(line), 10, 64)
		}

		reader := bufio.NewReaderSize(file, 4096)

		fmt.Println("Opened file " + buffer.String())

		var linesRead uint64 = 0

		startTime := time.Now()

		for {
			if linesRead > 0 && linesRead == params.lineLimit {
				log.Println("Reached line limit, concluding analysis")
				break
			}

			line := recurseBuildCompleteLine(reader)
			if line == nil {
				fmt.Println("Lines: " + strconv.FormatUint(linesRead, 10))
				log.Println("Encountered error; concluding analysis")
				break
			}
			linesRead++

			if meetsCriteria(string(line), params.stringORQueries, params.stringANDQueries) &&
				meetsIntCriteria(line, params.intORQueries, params.intANDQueries) {
				parsed := getCommentDataFromLine(line, commentFields)
				relevantComments = append(relevantComments, parsed)
			}

			if linesRead%HundredThousand == 0 {
				if totalLines != 0 {
					percentDone := (float64(linesRead) / float64(totalLines)) * 100.0
					fmt.Println(strconv.FormatFloat(percentDone, 'f', 2, 64) + "%")
				} else {
					fmt.Println(strconv.FormatUint(linesRead, 10))
				}

			}
		}
		dif := time.Now().Sub(startTime).String()
		fmt.Println("Took " + dif + " to search " + strconv.FormatUint(linesRead, 10) + " comments of file " + buffer.String())
	}
	return relevantComments
}
