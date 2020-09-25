package s1

import (
	"reflect"
	"testing"
	"time"
)

func TestValueDate(t *testing.T) {
	date := "22.09.2020"
	got, err := value(date)
	if err != nil {
		t.Errorf("err=%v", err)
	}
	want := time.Unix(1600732800, 0).UTC()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestValueDateTime(t *testing.T) {
	dateTime := "22.09.2020 10:00:00"
	got, err := value(dateTime)
	if err != nil {
		t.Errorf("err=%v", err)
	}
	want := time.Unix(1600768800, 0).UTC()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestValueInteger(t *testing.T) {
	integers := []struct {
		arg  string
		want int
	}{
		{"0", 0},
		{"99", 99},
		{"999", 999},
		{"9999", 9999},
		{"1000000", 1e6},
		{"1000000000", 1e9},
		{"1000000000000", 1e12},
	}

	for _, tt := range integers {
		got, err := value(tt.arg)
		if err != nil {
			t.Errorf("err=%v", err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("got %v, want %v", got, tt.want)
		}
	}
}

func TestValueDecimal(t *testing.T) {
	decimals := []struct {
		arg  string
		want float64
	}{
		{"0.00", 0.00},
		{"99.99", 99.99},
		{"123.123", 123.123},
		{"9999.9999", 9999.9999},
	}

	for _, tt := range decimals {
		got, err := value(tt.arg)
		if err != nil {
			t.Errorf("err=%v", err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("got %v, want %v", got, tt.want)
		}
	}
}

func TestValueString(t *testing.T) {
	strings := []string{"hello", "world"}

	for i, tt := range strings {
		got, err := value(tt)
		if err != nil {
			t.Errorf("err=%v", err)
		}
		want := strings[i]
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}
