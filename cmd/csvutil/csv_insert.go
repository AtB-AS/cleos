package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
)

type csvReader interface {
	Header() ([]string, error)
	Next() ([]string, error)
}

func insertCSV(ctx context.Context, r csvReader) error {
	var insertTmpl = template.Must(template.New("insertTmpl").Parse(`
INSERT into {{ .Name }} ({{ .Columns }})
VALUES ({{ .ValuePlaceholders }})`))
	var (
		user     = os.Getenv("DB_USER")
		port     = os.Getenv("DB_PORT")
		host     = os.Getenv("DB_HOST")
		password = os.Getenv("DB_PASSWORD")
		dbname   = os.Getenv("DB_NAME")
	)
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// Grab the CSV header row. The header is used to populate the insert
	// statement template and to generate value placeholders.
	header, err := r.Header()
	if err != nil {
		return err
	}
	placeholders := placeHoldersFor(header)

	// Populate and execute our template.
	var buf bytes.Buffer
	var data = struct {
		Name              string
		Columns           string
		ValuePlaceholders string
	}{"sales_transactions", columnsFor(header), placeholders}
	if err = insertTmpl.Execute(&buf, data); err != nil {
		return err
	}
	insertStmt := buf.String()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for {
		row, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, insertStmt, valuesFor(row)...); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func placeHoldersFor(in []string) string {
	var s strings.Builder
	for i := range in {
		s.WriteString(fmt.Sprintf("$%d", i+1))
		if i != len(in)-1 {
			s.WriteString(",")
		}
	}
	return s.String()
}

func columnsFor(hdr []string) string {
	return strings.Join(hdr, ",")
}

func valuesFor(row []string) []interface{} {
	var out = make([]interface{}, 0, len(row))

	for _, v := range row {
		i, err := parseInt(v)
		if err == nil {
			out = append(out, i)
			continue
		}

		f, err := parseFloat(v)
		if err == nil {
			out = append(out, f)
			continue
		}

		d, err := parseDateTime(v)
		if err == nil {
			out = append(out, d)
			continue
		}

		out = append(out, v)
	}

	return out
}
