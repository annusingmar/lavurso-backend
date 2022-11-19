package types

import (
	"errors"
	"strconv"
	"time"
)

var (
	ErrInvalidDateFormat = errors.New("invalid date format")
)

type Date struct {
	*time.Time
}

func ParseDate(s string) (*Date, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	return &Date{&t}, nil
}

func (d *Date) String() string {
	return d.Format("2006-01-02")
}

func (d *Date) UnmarshalJSON(b []byte) error {
	dateString, err := strconv.Unquote(string(b))
	if err != nil {
		// err if not a string
		return ErrInvalidDateFormat
	}

	if dateString == "" {
		d.Time = nil
		return nil
	}

	t, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		return ErrInvalidDateFormat
	}

	d.Time = &t
	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	var fd string

	if d.Time != nil {
		fd = d.Time.Format("2006-01-02")
	} else {
		return []byte("null"), nil
	}

	return []byte(strconv.Quote(fd)), nil
}
