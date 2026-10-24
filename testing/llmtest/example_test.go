package llmtest_test

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms/fake"
	"github.com/tmc/langchaingo/testing/llmtest"
)

func ExampleTestModel() {
	model := fake.NewFakeLLM([]string{"OK", "Grace"})
	if err := llmtest.TestModel(context.Background(), model, "multiturn"); err != nil {
		fmt.Println(err)
	}
	fmt.Println("conforms")
	// Output: conforms
}
