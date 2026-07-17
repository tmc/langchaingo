// Obsolete generator: the vertex package is now hand-maintained.
//
// It used to regenerate llms/googleai/vertex/vertex.go from googleai.go, but the
// two packages now target different SDKs (google.golang.org/genai for googleai vs
// cloud.google.com/go/vertexai/genai for vertex), so a mechanical rewrite of the
// old import no longer produces a compilable file. Running it would clobber the
// hand-written vertex code. It refuses to run rather than silently corrupt.
//
// nolint
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "generate-vertex is obsolete: the vertex package is hand-maintained.")
	fmt.Fprintln(os.Stderr, "googleai uses google.golang.org/genai while vertex uses cloud.google.com/go/vertexai/genai;")
	fmt.Fprintln(os.Stderr, "regenerating vertex.go from googleai.go would produce a file that does not compile.")
	fmt.Fprintln(os.Stderr, "Edit llms/googleai/vertex/vertex.go by hand. See llms/googleai/README.md.")
	os.Exit(1)
}
