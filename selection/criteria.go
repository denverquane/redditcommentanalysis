package selection

var AllMonths = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

var AllYears = []string{"2010", "2011", "2012", "2014", "2015", "2016"}

var commentFields = map[string]string{
	"author":    "str",
	"subreddit": "str",
	"body":      "str",
	"score":     "int",
}

var BasicSchema = commentSchema{
	name:   "Basic",
	schema: commentFields,
}

type commentSchema struct {
	name   string
	schema map[string]string
}
