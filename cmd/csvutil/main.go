package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"github.com/atb-as/cleos/pkg/cleos/postgres"
	"github.com/atb-as/cleos/pkg/cleos/s1"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
	}
	switch os.Args[1] {
	case "schema":
		args := os.Args[2:]
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "usage: %s %s tableName file.csv\n", os.Args[0], os.Args[1])
			os.Exit(1)
		}

		file, err := os.Open(args[1])
		if err != nil {
			return err
		}

		s1reader := s1.NewReader(file)

		hdr, err := s1reader.Header()
		if err != nil {
			return err
		}

		row1, err := s1reader.Row()
		if err != nil {
			return err
		}

		s, err := postgres.Schema(hdr, row1, args[0])
		if err != nil {
			return err
		}
		fmt.Println(s)
	case "insert":
		args := os.Args[2:]
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "usage: %s %s tableName file.csv", os.Args[0], os.Args[1])
			os.Exit(1)
		}
		file, err := os.Open(args[1])
		if err != nil {
			return err
		}

		reader := s1.NewReader(file)

		return insertCSV(context.Background(), reader, args[0])
	default:
		printUsage()
	}
	return nil
}

func printUsage() {
	fmt.Println(`invalid command, valid commands are:

	insert	insert values from csv file
	schema	generate sql schema from csv headers`)
	os.Exit(1)
}
