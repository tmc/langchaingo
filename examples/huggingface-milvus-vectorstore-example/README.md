# Milvus vector store with local embeddings via huggingface TEI.

Dependencies:
- [Text Embeddings Inference](https://github.com/huggingface/text-embeddings-inference) (optional)

```shell
# start milvus and text embeddings interface
docker-compose up -d


```

## For faster embeddings

```shell
# stop txt-inference container
docker compose stop txt-inference

# start embedding server
text-embeddings-router --model-id thenlper/gte-large --port 5500
```