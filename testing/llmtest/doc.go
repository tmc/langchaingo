// Package llmtest provides support for testing implementations of
// llms.Model, in the spirit of testing/fstest.
//
// # TestModel
//
// TestModel checks that a model satisfies the llms.Model contract. It
// is not coupled to *testing.T: it returns an error describing every
// violation found, so it can run anywhere:
//
//	func TestConformance(t *testing.T) {
//	    model, err := myprovider.New()
//	    if err != nil {
//	        t.Fatal(err)
//	    }
//	    if err := llmtest.TestModel(context.Background(), model, "streaming", "tools"); err != nil {
//	        t.Fatal(err)
//	    }
//	}
//
// The variadic list names the capabilities the model must demonstrate,
// as fstest.TestFS's expected files do; unnamed capabilities are not
// exercised. Against a live provider TestModel performs network calls;
// record them (for example with httprr) to run offline.
//
// The llms/fake package provides a canned-response model that conforms
// to the baseline contract, playing the role fstest.MapFS plays for
// io/fs.
//
// # TestLLM
//
// TestLLM is an older *testing.T-based harness that probes capabilities
// by issuing live requests. New tests should prefer TestModel.
package llmtest
