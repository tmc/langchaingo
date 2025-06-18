// Package compliance provides a test suite to verify provider implementations.
//
// The compliance suite tests that LLM providers correctly implement the
// standard interfaces and behave consistently across different implementations.
//
// Usage:
//
//	func TestProviderCompliance(t *testing.T) {
//	    model, err := provider.New()
//	    if err != nil {
//	        t.Fatal(err)
//	    }
//
//	    suite := compliance.NewSuite("provider", model)
//	    suite.Run(t)
//	}
package compliance
