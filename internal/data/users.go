package data

import (
	"context"
	"database/sql"
	"time"
)

const (
	Administrator = iota + 1
	Teacher
	Parent
	Student
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  []byte    `json:"-"`
	Phone     string    `json:"phone,omitempty"`
	Address   string    `json:"address,omitempty"`
	BirthDate time.Time `json:"birth_date,omitempty"`
	Role      int       `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	Version   int       `json:"version"`
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) AllUsers() ([]*User, error) {
	query := `SELECT id, name, email, password, phone, address, birth_date, role, created_at, version FROM users`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Password,
			&user.Phone,
			&user.Address,
			&user.BirthDate,
			&user.Role,
			&user.CreatedAt,
			&user.Version,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil

}
