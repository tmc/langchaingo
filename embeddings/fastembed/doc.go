// Package fastembed provides FastEmbed embeddings integration for LangChain Go.
//
// FastEmbed is a lightweight, fast embedding library that runs models locally using ONNX.
// It supports various pre-trained models including BGE and sentence-transformers models.
//
// The package supports both document and query embeddings with different prefixing strategies:
//   - Document embeddings can use "default" or "passage" prefixing
//   - Query embeddings automatically use "query:" prefixing for optimal search performance
//
// Example usage:
//
//	embedder, err := fastembed.NewFastEmbed(
//		fastembed.WithModel(fastembed.BGESmallENV15),
//		fastembed.WithCacheDir("./models"),
//		fastembed.WithBatchSize(128),
//		fastembed.WithDocEmbedType("passage"),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer embedder.Close()
//
//	// Embed documents
//	docs := []string{"This is a document", "Another document"}
//	docEmbeddings, err := embedder.EmbedDocuments(ctx, docs)
//
//	// Embed query
//	query := "search query"
//	queryEmbedding, err := embedder.EmbedQuery(ctx, query)
//
// The implementation automatically downloads and caches models on first use.
// Models are stored locally and run entirely offline after initial download.
package fastembed
