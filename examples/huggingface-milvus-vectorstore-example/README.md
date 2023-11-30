# Milvus vector store with local embeddings via huggingface TEI.

Dependencies:
- [Text Embeddings Inference](https://github.com/huggingface/text-embeddings-inference)
- [Ollama](https://ollama.ai/) 

```shell
# start milvus
docker-compose up -d

#start embedding server
text-embeddings-router --model-id thenlper/gte-large --port 5500
```
