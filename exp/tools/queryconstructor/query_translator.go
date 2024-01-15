package queryconstructor

type QueryTranslator interface {
	Comparator()
	Comparison()
	Operator()
	Operation()
	StructuredQuery()
}
