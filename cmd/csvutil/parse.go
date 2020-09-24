package main

import (
	"fmt"
	"strconv"
	"time"
)

func parseFloat(s string) (float32, error) {
	f, err := strconv.ParseFloat(s, 32)
	return float32(f), err
}

func parseInt(s string) (int, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	return int(i), err
}

func parseDateTime(s string) (time.Time, error) {
	var dateLayouts = []string{
		"02.01.2006",
		"02.01.2006 15:04:05",
	}
	for _, d := range dateLayouts {
		t, err := time.Parse(d, s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date")
}
