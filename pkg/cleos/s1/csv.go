package s1

import (
	"encoding/csv"
	"io"
	"regexp"
	"strconv"
	"time"
)

const (
	Separator = ';'
)

type Reader struct {
	hdrRead bool
	hdr     []string
	csv     *csv.Reader
}

func NewReader(r io.Reader) *Reader {
	reader := &Reader{
		csv: csv.NewReader(r),
	}
	reader.csv.Comma = Separator

	return reader
}

func (r *Reader) Header() ([]string, error) {
	if !r.hdrRead {
		return r.readHeader()
	}
	return r.hdr, nil
}

func (r *Reader) Row() ([]interface{}, error) {
	row, err := r.csv.Read()
	if err != nil {
		return nil, err
	}

	var ret = make([]interface{}, 0, len(row))
	for _, v := range row {
		va, err := value(v)
		if err != nil {
			return nil, err
		}
		ret = append(ret, va)
	}

	return ret, nil
}

// value converts a string value present in a column in an S-1 report to its
// corresponding Go value.
//
// Types in the S-1 report as documented in
// "M09-12 Generation of Sales Documentation and General Ledger Data":
//
// Date: A calendar date without time: dd.MM.yyyy
// DateTime: A calendar date with a time component: dd.MM.yyyy HH:mm:ss
// Timestamp: An unrestricted point in time, typically a time of creation. Not
// seen in the wild yet, so not included.
// Integer: A number with no decimals.
// Amount: A number always formatted with 2 decimals. NOK currency.
// Decimal: A number with at least one decimal.
// String: A quoted ("") value. Go's encoding/csv package trims the quotes for
// us automatically.
func value(v string) (interface{}, error) {
	amount := regexp.MustCompile("^\\d+\\.\\d{2}$")
	date := regexp.MustCompile("^\\d{2}\\.\\d{2}\\.\\d{4}$")
	dateTime := regexp.MustCompile("^\\d{2}\\.\\d{2}\\.\\d{4}\\s+\\d{2}:\\d{2}:\\d{2}$")
	decimal := regexp.MustCompile("^\\d+\\.\\d+$")
	integer := regexp.MustCompile("^\\d+$")

	switch {
	case date.MatchString(v):
		return decodeDate(v)
	case dateTime.MatchString(v):
		return decodeDateTime(v)
	case integer.MatchString(v):
		return decodeInt(v)
	case amount.MatchString(v):
		fallthrough
	case decimal.MatchString(v):
		return decodeDecimal(v)
	default:
		return v, nil
	}
}

func decodeDecimal(v string) (float64, error) {
	return strconv.ParseFloat(v, 64)
}

func decodeInt(v string) (int, error) {
	return strconv.Atoi(v)
}

func decodeDateTime(v string) (time.Time, error) {
	const layout = "02.01.2006 15:04:05"
	return time.Parse(layout, v)
}

func decodeDate(v string) (time.Time, error) {
	const layout = "02.01.2006"
	return time.Parse(layout, v)
}

func (r *Reader) readHeader() ([]string, error) {
	hdr, err := r.csv.Read()
	if err != nil {
		return nil, err
	}
	r.hdr = hdr
	r.hdrRead = true

	return hdr, nil
}
