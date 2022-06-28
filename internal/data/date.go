package data

import (
	"errors"
	"strconv"
	"time"
)

var (
	ErrInvalidDateFormat = errors.New("invalid date format")
)

type Date struct {
	time.Time
}

func (d *Date) UnmarshalJSON(b []byte) error {
	dateString, err := strconv.Unquote(string(b))
	if err != nil {
		// err if not a string
		return ErrInvalidDateFormat
	}

	t, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		return ErrInvalidDateFormat
	}

	d.Time = t
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	fd := d.Time.Format("2006-01-02")
	return []byte(strconv.Quote(fd)), nil
}
