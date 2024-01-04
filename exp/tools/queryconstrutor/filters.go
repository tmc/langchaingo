package queryconstrutor

// Enumerator of the operations.
type Operator = string

const (
	OperatorAnd Operator = "and"
	OperatorOr  Operator = "or"
	OperatorNot Operator = "not"
)

// Enumerator of the comparison operators.
type Comparator = string

const (
	ComparatorEQ      Comparator = "eq"
	ComparatorNE      Comparator = "ne"
	ComparatorGT      Comparator = "gt"
	ComparatorGTE     Comparator = "gte"
	ComparatorLT      Comparator = "lt"
	ComparatorLTE     Comparator = "lte"
	ComparatorCONTAIN Comparator = "contain"
	ComparatorLIKE    Comparator = "like"
	ComparatorIN      Comparator = "in"
	ComparatorNIN     Comparator = "nin"
)
