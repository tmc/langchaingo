# pgvector store with OpenAI embeddings

Start Postgres:

```shell
docker-compose up -d
```

`postgres.Dockerfile` extends the official Postgres image and installs the pgvector extension.

`create_extension.sql` enables the pgvector extension and should run automatically when the container starts for the
first time.

Run the example:

```shell
export OPENAI_API_KEY=<your key>
go run pgvector_vectorstore_example.go
```

