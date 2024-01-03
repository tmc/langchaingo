package queryconstrutor

type AttributeInfo struct {
	Name        string
	Description string
	Type        string
}

type Example struct {
	Input string
	Ouput interface{}
}
