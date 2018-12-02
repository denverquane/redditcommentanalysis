package selection

var AllMonths = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

var AllYears = []string{"2010", "2011", "2012", "2014", "2015", "2016"}

//NOTE changing this datamodel can invalidate older or future data captures...
var commentFields = map[string]string{
	"author":    "str",
	"subreddit": "str",
	"body":      "str",
	"score":     "int",
}

//var BasicSchema = commentSchema{
//	name:   "Basic",
//	schema: commentFields,
//}
var BestSchema = commentSchema{
	name:   "Best",
	schema: commentFields,
}

type commentSchema struct {
	name   string
	schema map[string]string
}

type Criteria struct {
	Test  string
	Value string
}
