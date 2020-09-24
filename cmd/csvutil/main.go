package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
	}
	switch os.Args[1] {
	case "schema":
		args := os.Args[2:]
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "usage: %s %s file.csv\n", os.Args[0], os.Args[1])
			os.Exit(1)
		}
		file, err := os.Open(args[0])
		if err != nil {
			return err
		}

		reader := csv.NewReader(file)
		reader.Comma = ';'

		s, err := createSchema(reader)
		if err != nil {
			return err
		}
		fmt.Println(s)
	case "insert":
		args := os.Args[2:]
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "usage: %s %s file.csv", os.Args[0], os.Args[1])
			os.Exit(1)
		}
		file, err := os.Open(args[0])
		if err != nil {
			return err
		}

		r := csv.NewReader(file)
		r.Comma = ';'

		reader := &csvReaderImpl{r: r}

		return insertCSV(context.Background(), reader)
	default:
		printUsage()
	}
	return nil
}

func printUsage() {
	fmt.Println(`invalid command, valid commands are:

	schema	generate sql schema from csv headers`)
	os.Exit(1)
}
