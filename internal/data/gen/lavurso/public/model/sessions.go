//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"github.com/annusingmar/lavurso-backend/internal/types"
	"time"
)

type Sessions struct {
	ID           int          `sql:"primary_key" json:"id,omitempty"`
	Token        *types.Token `json:"token"`
	UserID       *int         `json:"user_id,omitempty"`
	Expires      *time.Time   `json:"expires,omitempty"`
	LoginIP      *string      `json:"login_ip,omitempty"`
	LoginBrowser *string      `json:"login_browser,omitempty"`
	LoggedIn     *time.Time   `json:"logged_in,omitempty"`
	LastSeen     *time.Time   `json:"last_seen,omitempty"`
}
