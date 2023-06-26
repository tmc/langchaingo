package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools/sql_database"
	_ "github.com/tmc/langchaingo/tools/sql_database/sqlite3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func makeSample(dsn string) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	create table foo (id integer not null primary key, name text);
	delete from foo;
	create table foo1 (id integer not null primary key, name text);
	delete from foo1;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into foo(id, name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("Index %03d", i))
		if err != nil {
			log.Fatal(err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

}
func run() error {
	llm, err := openai.New()
	if err != nil {
		return err
	}

	const dsn = "./foo.db"
	os.Remove(dsn)
	defer os.Remove(dsn)

	makeSample(dsn)

	db, err := sql_database.NewSQLDatabaseWithDSN("sqlite3", dsn, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	sqlDatabaseChain := chains.NewSQLDatabaseChain(llm, 10, db)
	ctx := context.Background()
	out, err := chains.Run(ctx, sqlDatabaseChain, "Return all data from the foo table where the ID is less than 23.")
	fmt.Println(out)

	out, err = chains.Run(ctx, sqlDatabaseChain, "How many data entries are there in the foo table?")
	fmt.Println(out)
	return err
}
