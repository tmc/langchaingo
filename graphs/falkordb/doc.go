// Package falkordb provides a FalkorDB implementation of the graphs.GraphStore interface.
//
// FalkorDB is a super fast Graph Database that uses GraphBLAS under the hood for its
// sparse adjacency matrix graph representation. This package enables you to connect
// to FalkorDB instances and perform graph operations using the standard GraphStore
// interface.
//
// Basic usage:
//
//	db, err := falkordb.New("my_graph",
//	    falkordb.WithHost("localhost"),
//	    falkordb.WithPort(6379),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Add graph documents
//	err = db.AddGraphDocuments(ctx, graphDocs)
//
// Security note: Make sure that the database connection uses credentials
// that are narrowly-scoped to only include necessary permissions.
// Failure to do so may result in data corruption or loss.
package falkordb
