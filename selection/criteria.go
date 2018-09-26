package selection

import (
	"github.com/valyala/fastjson"
	"strings"
)

var AllMonths = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

type intVsInt func(int, int) bool

func IsGreater(x int, y int) bool {
	return x > y
}

func IsEqual(x int, y int) bool {
	return x == y
}

func IsLess(x int, y int) bool {
	return x < y
}

type IntCriteria struct {
	key       string
	reference int
	pred      intVsInt
}

func meetsCriteria(line string, ORs []string, ANDs []string) bool {
	var metCriteria bool
	if len(ORs) > 0 {
		metCriteria = false
	} else {
		metCriteria = true //no ORs to test- automatically valid until AND step
	}

	for _, v := range ORs {
		if strings.Contains(strings.ToLower(line), strings.ToLower(v)) {
			metCriteria = true //met a SINGLE one of the OR criteria; continues on
			break
		}
	}
	if !metCriteria { //failed all the ORs
		return false
	}

	for _, v := range ANDs {
		if !strings.Contains(strings.ToLower(line), strings.ToLower(v)) {
			return false
		}
	}
	return true
}

func meetsIntCriteria(line []byte, ORs []IntCriteria, ANDs []IntCriteria) bool {
	var metCriteria bool
	if len(ORs) > 0 {
		metCriteria = false
	} else {
		metCriteria = true //no ORs to test- automatically valid until AND step
	}

	for _, v := range ORs {
		y := fastjson.GetInt(line, v.key)
		if v.pred(v.reference, y) {
			metCriteria = true //met a SINGLE one of the OR criteria; continues on
			break
		}
	}
	if !metCriteria { //failed all the ORs
		return false
	}

	for _, v := range ANDs {
		y := fastjson.GetInt(line, v.key)
		if !v.pred(v.reference, y) {
			return false //failed a single "AND"
		}
	}
	return true
}
