package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools/sqldatabase"
	_ "github.com/tmc/langchaingo/tools/sqldatabase/sqlite3"
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
		_, err = stmt.Exec(i, fmt.Sprintf("Foo %03d", i))
		if err != nil {
			log.Fatal(err)
		}
	}

	stmt1, err := tx.Prepare("insert into foo1(id, name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt1.Close()
	for i := 0; i < 200; i++ {
		_, err = stmt1.Exec(i, fmt.Sprintf("Foo1 %03d", i))
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

	db, err := sqldatabase.NewSQLDatabaseWithDSN("sqlite3", dsn, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	sqlDatabaseChain := chains.NewSQLDatabaseChain(llm, 100, db)
	ctx := context.Background()
	out, err := chains.Run(ctx, sqlDatabaseChain, "Return all rows from the foo table where the ID is less than 23.")
	if err != nil {
		return err
	}
	fmt.Println(out)

	input := map[string]any{
		"query":              "Return all rows that the ID is less than 23.",
		"table_names_to_use": []string{"foo"},
	}
	out, err = chains.Predict(ctx, sqlDatabaseChain, input)
	if err != nil {
		return err
	}
	fmt.Println(out)

	out, err = chains.Run(ctx, sqlDatabaseChain, "Which table has more data, foo or foo1?")
	if err != nil {
		return err
	}
	fmt.Println(out)
	return err
}
