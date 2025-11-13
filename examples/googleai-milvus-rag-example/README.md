# Google Gemini + Milvus RAG Example

Welcome to the **Google Gemini + Milvus RAG Example**!  
This Go example demonstrates how to build a simple **Retrieval-Augmented Generation (RAG)** pipeline using the [`langchaingo`](https://github.com/tmc/langchaingo) library, **Google Gemini** for embeddings and text generation, and **Milvus** as a vector database.

This example shows how you can combine vector search with large language models to answer questions based on your own documents — a foundation for building intelligent knowledge assistants, search tools, and chatbots.

---

## 🧩 What This Example Does

This example program:

1. Connects to the **Google Gemini API** using your API key.
2. Connects to a **Milvus** instance (local or cloud).
3. Embeds and inserts a few sample documents into Milvus.
4. Performs a **semantic similarity search** for a given query.
5. Sends the retrieved information to Gemini for a natural-language answer.

---

## ⚙️ How It Works

1. **Environment Setup**
   - Loads the following environment variables:
     - `GEMINI_API_KEY` → Required. Your Google Gemini API key.
     - `MILVUS_URL` → Required. Address of your Milvus instance (`localhost:19530` for local Milvus, or your cloud endpoint).
     - `MILVUS_API_KEY` → Optional. Only required for Milvus cloud instances.

2. **Embedding & Storage**
   - The program uses Gemini’s embedding model (`gemini-embedding-001`) to convert text into vectors.
   - These vectors are stored in Milvus under the collection name `"docs"`.

3. **Retrieval**
   - A similarity search is performed for the query `"What is machine learning?"`.
   - Milvus returns the top documents with the highest similarity scores.

4. **Answer Generation**
   - The relevant text is provided to Gemini (`gemini-2.0-flash`) to generate a clear, concise English answer.

---

## 🪄 Example Documents

The following example notes are stored in Milvus:

- “The capital of France is Paris. It is known for its culture, art, and cuisine.”  
- “Machine learning is a branch of artificial intelligence that enables systems to learn from data and improve over time without being explicitly programmed.”  
- “Photosynthesis is the process by which green plants use sunlight to synthesize nutrients from carbon dioxide and water.”

---

## 🚀 Running the Example

### 1. Requirements

Make sure you have:

- **Go 1.25+** installed  
- Access to a **Google Gemini API key**  
- Access to a **Milvus instance** (local or cloud)

Verify your Go installation:

```bash
go version
```
If you don’t have Milvus running locally, you can start it quickly with Docker:
```bash
docker run -p 19530:19530 milvusdb/milvus:latest
```
### 2. Clone and Navigate
```bash
git clone https://github.com/tmc/langchaingo.git
cd langchaingo/examples/vectorstores/milvus_gemini_rag
```
### 3. Set Environment Variables
Linux / macOS
```bash
export GEMINI_API_KEY=your_gemini_api_key_here
export MILVUS_URL=localhost:19530
# Optional for Milvus Cloud
export MILVUS_API_KEY=your_milvus_api_key_here
```
Window
```bash
$env:GEMINI_API_KEY="your_gemini_api_key_here"
$env:MILVUS_URL="localhost:19530"
$env:MILVUS_API_KEY="your_milvus_api_key_here"
```
### 4. Run the Program
```bash
go run main.go
```
### 5. Example Output
```bash
Answer: Machine learning is a branch of artificial intelligence that allows systems to learn from data and improve automatically through experience.
```