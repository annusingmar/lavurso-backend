package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
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
	var ds string
	if err := json.Unmarshal(b, &ds); err != nil {
		return ErrInvalidDateFormat
	}

	if ds == "" {
		d.Time = nil
		return nil
	}

	date, err := ParseDate(ds)
	if err != nil {
		return err
	}

	*d = *date
	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	var fd string

	if d.Time != nil {
		fd = d.String()
	} else {
		return []byte("null"), nil
	}

	return json.Marshal(fd)
}

func (d *Date) Scan(src any) error {
	date := src.(time.Time)
	d.Time = &date
	return nil
}

func (d *Date) Value() (driver.Value, error) {
	return *d.Time, nil
}
