# 🌟 Milvus Vector Store with Local Embeddings Example 🚀

Welcome to this exciting example showcasing the power of Milvus vector store combined with local embeddings using Hugging Face's Text Embeddings Inference (TEI)! 🎉

## 🧠 What Does This Example Do?

This Go program demonstrates how to:

1. Set up a Milvus vector store
2. Use local embeddings with Hugging Face's TEI
3. Add documents about cities to the vector store
4. Perform similarity searches on the stored data

The example focuses on storing information about various cities worldwide and then querying this data to find cities based on different criteria.

## 🏙️ City Data

We store information about cities including:
- Tokyo, Kyoto, Hiroshima, and other Japanese cities
- European cities like Paris and London
- South American cities like Santiago, Buenos Aires, and Rio de Janeiro

Each city entry includes:
- City name
- Population
- Area

## 🔍 Similarity Searches

The program runs three example searches:

1. **Up to 5 Cities in Japan**: Finds Japanese cities in the dataset
2. **A City in South America**: Locates a South American city
3. **Cities in Europe**: Searches for European cities

## 🚀 How It Works

1. The program sets up a Milvus vector store with local embeddings
2. City data is added to the vector store
3. Similarity searches are performed using natural language queries
4. Results are displayed, showing matching cities with their population and area

## 🛠️ Setup

To run this example, make sure you have:

1. Docker and Docker Compose installed
2. Go environment set up
3. Milvus and Text Embeddings Inference services running (use the provided docker-compose.yml)

## 🎈 Have Fun!

Enjoy exploring this example and learning about vector stores, embeddings, and similarity searches! Feel free to modify the queries or add more cities to see how the results change. Happy coding! 😄
