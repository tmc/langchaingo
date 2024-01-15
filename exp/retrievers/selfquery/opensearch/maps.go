package selfquery_opensearch

import "github.com/tmc/langchaingo/exp/tools/queryconstructor"

var ComparatorMap map[string]string = map[string]string{
	queryconstructor.ComparatorEQ:      "term",
	queryconstructor.ComparatorLT:      "lt",
	queryconstructor.ComparatorLTE:     "lte",
	queryconstructor.ComparatorGT:      "gt",
	queryconstructor.ComparatorGTE:     "gte",
	queryconstructor.ComparatorCONTAIN: "match",
	queryconstructor.ComparatorLIKE:    "fuzzy",
}

var OperatorMap map[string]string = map[string]string{
	queryconstructor.OperatorAnd: "must",
	queryconstructor.OperatorOr:  "should",
	queryconstructor.OperatorNot: "must_not",
}
