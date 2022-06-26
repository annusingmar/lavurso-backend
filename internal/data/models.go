package data

import "database/sql"

type Models struct {
	Users    UserModel
	Classes  ClassModel
	Subjects SubjectModel
}

func NewModel(db *sql.DB) Models {
	return Models{
		Users:    UserModel{DB: db},
		Classes:  ClassModel{DB: db},
		Subjects: SubjectModel{DB: db},
	}
}
