package aplos

import (
	"testing"
	"time"
)

func d(year int, month time.Month, day int) Date {
	return Date{Year: year, Month: month, Day: day}
}

func TestDateString(t *testing.T) {
	tests := []struct {
		in   Date
		want string
	}{
		{
			in:   d(2020, time.January, 2),
			want: "2020-01-02",
		},
		{
			in:   d(1993, time.October, 10),
			want: "1993-10-10",
		},
		{
			in:   d(999, time.December, 31),
			want: "0999-12-31",
		},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			got := test.in.String()
			if got != test.want {
				t.Errorf("Date.String() = %q, want %q", got, test.want)
			}
		})
	}
}
