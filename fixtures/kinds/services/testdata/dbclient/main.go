package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/jackc/pgx/v5"
)

func run() (runErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dsn := ""
	if len(os.Args) != 2 {
		return fmt.Errorf("expected env or dbtest mode")
	}
	switch os.Args[1] {
	case "env":
	case "dbtest":
		database, cleanup, err := dbtest.Open(dbtest.Options{Name: "fixture-child"})
		if err != nil {
			return err
		}
		defer func() { runErr = errors.Join(runErr, cleanup()) }()
		dsn = database.DSN()
	default:
		return fmt.Errorf("unknown mode %q", os.Args[1])
	}
	connection, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer func() { runErr = errors.Join(runErr, connection.Close(ctx)) }()
	var database string
	if err := connection.QueryRow(ctx, "SELECT current_database()").Scan(&database); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]string{"database": database})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
