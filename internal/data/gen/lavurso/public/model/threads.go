//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"time"
)

type Threads struct {
	ID        int        `sql:"primary_key" json:"id,omitempty"`
	UserID    *int       `json:"user_id,omitempty"`
	Title     *string    `json:"title,omitempty"`
	Locked    *bool      `json:"locked,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
