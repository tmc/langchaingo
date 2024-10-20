# Ollama Retrieval Example

This example demonstrates how to use retrieval capabilities with the Ollama language model using the langchaingo library. It showcases a simple retrieval system got the answer from you provided context.

## What it does

1. It sets up an Ollama language model instance using the "mistral" model.
2. Providing specific information or domain knowledge ensures that the AI gives priority to this when responding.
3. The main questions we will ask are "What is this?" If there is direct relevance, prioritize answering based on the information we provided. Otherwise, simply respond without referencing our information, and avoid mentioning the provided details unless necessary.
4. When asking the AI a question, we expect `foo` and `f4` to be answered based on the information we provided.
5. When asking about `foodpanda` and `panda`, we expect the AI to respond based on its own knowledge.

## How to Run

1. Ensure you have Ollama set up and running on your system.
2. Run the example with: `go run ollama_retrieval_example.go`.

## What to Expect

When you run this program, the AI will respond to your questions asynchronously. If there is any information you have provided for reference, it will prioritize that in its response. Otherwise, the AI will answer based on its own knowledge.

This is an example executed on my laptop.
```bash
#2 foo
9527

#1 foodpanda
Foodpanda is an online food delivery service platform. It allows users to order food from various restaurants in their area and have it delivered to their doorstep.

#3 f4
F4 refers to a Taiwanese boy band that was formed in 2001 after the success of the Taiwanese drama Meteor Garden, where they starred.

#4 panda
Panda is a type of bear native to China. It is known for its black and white fur. The term "panda" is not directly related to the provided context about Foo or F4.
```

## Note

I am using the mistral model, and different models may provide varying or even unexpected responses. Some may not consider the fact that I provided `foo` as meaning `9527`, while others might heavily rely on the information I supplied. Since I didn’t provide any information about what `food` is, the AI might respond with `"I don’t know."` Optimizing the prompt or refining the provided information may help improve this issue.
