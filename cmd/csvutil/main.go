package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
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

func createSchema(r *csv.Reader) (string, error) {
	header, err := r.Read()
	if err != nil {
		return "", err
	}

	row1, err := r.Read()
	if err != nil {
		return "", err
	}

	s := &strings.Builder{}
	writeComma := func(v bool) {
		if !v {
			s.Write([]byte(",\n"))
		}
	}
	writeSchemaHeader(s, "sales_transactions")
	for i, column := range row1 {
		hdr := header[i]
		last := i == len(header)-1

		_, err = parseInt(column)
		if err == nil {
			writeIntColumn(s, hdr)
			writeComma(last)
			continue
		}

		_, err := parseFloat(column)
		if err == nil {
			writeFloatColumn(s, hdr)
			writeComma(last)
			continue
		}

		_, err = parseDate(column)
		if err == nil {
			writeDateColumn(s, hdr)
			writeComma(last)
			continue
		}

		writeStringColumn(s, hdr, row1[i])
		writeComma(last)
	}
	writeSchemaFooter(s)

	return s.String(), nil
}

func writeStringColumn(w io.Writer, hdr string, v string) {
	sz := 255
	for {
		if len(v) < sz {
			break
		}
		sz += sz
	}
	w.Write([]byte(fmt.Sprintf("%s varchar(%d)", strings.ToLower(hdr), sz)))
}

func writeDateColumn(w io.Writer, hdr string) {
	w.Write([]byte(fmt.Sprintf("%s date", strings.ToLower(hdr))))
}

func writeIntColumn(w io.Writer, hdr string) {
	w.Write([]byte(fmt.Sprintf("%s int", strings.ToLower(hdr))))
}

func writeFloatColumn(w io.Writer, hdr string) {
	w.Write([]byte(fmt.Sprintf("%s numeric", strings.ToLower(hdr))))
}

func writeSchemaFooter(w io.Writer) {
	w.Write([]byte("\n);"))
}

func writeSchemaHeader(w io.Writer, tableName string) {
	w.Write([]byte(fmt.Sprintf("CREATE TABLE %s (\n", tableName)))
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func parseDate(s string) (time.Time, error) {
	var dateLayouts = []string{
		"02.01.2006",
	}
	for _, d := range dateLayouts {
		t, err := time.Parse(d, s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date")
}
