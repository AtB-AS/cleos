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
	Row() ([]interface{}, error)
}

func insertCSV(ctx context.Context, r csvReader, tableName string) error {
	var insertTmpl = template.Must(template.New("insertTmpl").Parse(`
INSERT into {{ .Name }} ({{ .Columns }})
VALUES ({{ .Placeholders }})`))
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

	// Populate and execute our template.
	var buf bytes.Buffer
	var data = struct {
		Name         string
		Columns      string
		Placeholders string
	}{tableName, columns(header), placeholders(header)}
	if err = insertTmpl.Execute(&buf, data); err != nil {
		return err
	}
	insertStmt := buf.String()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() error {
		err = tx.Rollback()
		return err
	}()

	for {
		row, err := r.Row()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, insertStmt, row...); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func placeholders(in []string) string {
	var s strings.Builder
	for i := range in {
		s.WriteString(fmt.Sprintf("$%d", i+1))
		if i != len(in)-1 {
			s.WriteString(",")
		}
	}
	return s.String()
}

func columns(hdr []string) string {
	return strings.Join(hdr, ",")
}
