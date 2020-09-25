package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func createSchema(r csvReader, tableName string) (string, error) {
	header, err := r.Header()
	if err != nil {
		return "", err
	}

	row1, err := r.Row()
	if err != nil {
		return "", err
	}

	s := &strings.Builder{}
	writeComma := func(v bool) {
		if !v {
			s.Write([]byte(",\n"))
		}
	}

	writeSchemaHeader(s, tableName)
	for i, val := range row1 {
		last := i == len(row1)-1
		colName := header[i]
		switch t := val.(type) {
		case string:
			writeStringColumn(s, colName, len(t))
		case float64:
			writeFloatColumn(s, colName)
		case time.Time:
			writeDateColumn(s, colName)
		case int:
			writeIntColumn(s, colName, t)
		}
		writeComma(last)
	}
	writeSchemaFooter(s)

	return s.String(), nil
}

func writeStringColumn(s *strings.Builder, colName string, length int) {
	sz := 255
	for {
		if length < sz {
			break
		}
		sz += sz
	}
	s.WriteString(fmt.Sprintf("%s varchar(%d)", strings.ToLower(colName), sz))
}

func writeDateColumn(s *strings.Builder, colName string) {
	// Naive as we are, this is probably good enough for this use case.
	if strings.Contains(strings.ToLower(colName), "time") {
		s.WriteString(fmt.Sprintf("%s time", strings.ToLower(colName)))
		return
	}
	s.WriteString(fmt.Sprintf("%s date", strings.ToLower(colName)))
}

func writeIntColumn(s *strings.Builder, colName string, t int) {
	if t > math.MaxInt32 {
		s.WriteString(fmt.Sprintf("%s bigint", strings.ToLower(colName)))
		return
	}
	s.WriteString(fmt.Sprintf("%s int", strings.ToLower(colName)))
}

func writeFloatColumn(s *strings.Builder, colName string) {
	s.WriteString(fmt.Sprintf("%s numeric", strings.ToLower(colName)))
}

func writeSchemaFooter(s *strings.Builder) {
	s.WriteString("\n);")
}

func writeSchemaHeader(s *strings.Builder, tableName string) {
	s.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", tableName))
}
