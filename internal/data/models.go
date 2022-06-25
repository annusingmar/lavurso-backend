package data

import "database/sql"

type Models struct {
	Users UserModel
}

func NewModel(db *sql.DB) Models {
	return Models{
		Users: UserModel{DB: db},
	}
}
