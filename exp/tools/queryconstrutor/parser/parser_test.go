package queryconstructor_parser

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	result := Parse("and(or(eq(\"artist\", true), eq(\"artist\", \"Katy Perry\")), lt(\"length\", 180), eq(\"genre\", \"pop\"))")
	// queryconstructor_parser.Parse(`and("test")`)
	fmt.Printf("result: %+v\n", result)

	for i, v := range (*result).args {
		fmt.Printf("i: %v\n", i)
		fmt.Printf("v: %v\n", v)
	}

	t.Fail()
}
