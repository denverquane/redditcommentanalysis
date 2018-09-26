package main

import (
	"fmt"
	"github.com/denverquane/redditcommentanalysis/selection"
	"strconv"
)

const BaseDataDirectory = "D:/Reddit_Data"
const SSDBaseDataDirectory = "C:/Users/Denver"

func main() {
	searchCriteria := selection.MakeSimpleSearchParams("2016", "Jan")
	searchCriteria.AddANDCriteria("author", "TuckRaker")
	commentData := selection.OpenAndSearchFile(searchCriteria, BaseDataDirectory)
	fmt.Println("Total comments matching criteria: " + strconv.Itoa(len(commentData)))
}
