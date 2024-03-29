package data

import "database/sql"

type Models struct {
	Users       UserModel
	Classes     ClassModel
	Subjects    SubjectModel
	Journals    JournalModel
	Lessons     LessonModel
	Assignments AssignmentModel
	Grades      GradeModel
	Marks       MarkModel
	Absences    AbsenceModel
	Groups      GroupModel
	Messaging   MessagingModel
	Sessions    SessionModel
	Years       YearModel
	Logs        LogModel
}

func NewModel(db *sql.DB) Models {
	return Models{
		Users:       UserModel{DB: db},
		Classes:     ClassModel{DB: db},
		Subjects:    SubjectModel{DB: db},
		Journals:    JournalModel{DB: db},
		Lessons:     LessonModel{DB: db},
		Assignments: AssignmentModel{DB: db},
		Grades:      GradeModel{DB: db},
		Marks:       MarkModel{DB: db},
		Absences:    AbsenceModel{DB: db},
		Groups:      GroupModel{DB: db},
		Messaging:   MessagingModel{DB: db},
		Sessions:    SessionModel{DB: db},
		Years:       YearModel{DB: db},
		Logs:        LogModel{DB: db},
	}
}
