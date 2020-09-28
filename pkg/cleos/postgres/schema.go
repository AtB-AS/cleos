package postgres

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Schema returns a SQL statement for creating a table for the values contained
// in data. columns specifies column names. data and columns must have the same
// length.
func Schema(columns []string, data []interface{}, tableName string) (string, error) {
	if len(columns) != len(data) {
		return "", fmt.Errorf("columns and data data have different lengths")
	}
	s := &strings.Builder{}

	writeSchemaHeader(s, tableName)
	for i, val := range data {
		colName := columns[i]
		switch t := val.(type) {
		case float64:
			writeFloatColumn(s, colName)
		case time.Time:
			writeDateColumn(s, colName)
		case int:
			writeIntColumn(s, colName, t)
		case string:
			writeStringColumn(s, colName)
		default:
			return "", fmt.Errorf("unrecognized type for column %d", i)
		}
		if i != len(data)-1 {
			writeComma(s)
		}
	}
	writeSchemaFooter(s)

	return s.String(), nil
}

func writeComma(s *strings.Builder) {
	s.Write([]byte(",\n"))
}

func writeStringColumn(s *strings.Builder, colName string) {
	s.WriteString(fmt.Sprintf("%s text", strings.ToLower(colName)))
}

func writeDateColumn(s *strings.Builder, colName string) {
	// Naive as we are, this is probably good enough for our use case.
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
	s.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", tableName))
}
