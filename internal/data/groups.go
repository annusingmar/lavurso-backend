package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrNoSuchGroup        = errors.New("no such group")
	ErrUserAlreadyInGroup = errors.New("user already in group")
	ErrUserNotInGroup     = errors.New("user not in group")
)

type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type GroupModel struct {
	DB *pgxpool.Pool
}

// todo: insert, delete, update, add user to group, remove user from group

func (m GroupModel) GetGroupByID(groupID int) (*Group, error) {
	query := `SELECT id, name
	FROM groups
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var group Group

	err := m.DB.QueryRow(ctx, query, groupID).Scan(
		&group.ID,
		&group.Name,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchGroup
		default:
			return nil, err
		}
	}

	return &group, nil
}

func (m GroupModel) GetAllGroups() ([]*Group, error) {
	query := `SELECT id, name
	FROM groups`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var groups []*Group

	rows, err := m.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var group Group

		err = rows.Scan(
			&group.ID,
			&group.Name,
		)
		if err != nil {
			return nil, err
		}

		groups = append(groups, &group)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (m GroupModel) GetUsersByGroupID(groupID int) ([]*User, error) {
	query := `SELECT u.id, u.name, u.email, u.password, u.role, u.created_at, u.active, u.version
	FROM users_groups ug
	INNER JOIN users u
	ON ug.user_id = u.id
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query)
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
			&user.Password.Hashed,
			&user.Role,
			&user.CreatedAt,
			&user.Active,
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

func (m GroupModel) GetGroupsByUserID(userID int) ([]*Group, error) {
	query := `SELECT g.id, g.name
	FROM groups g
	INNER JOIN users_groups ug
	ON g.id = ug.group_id
	WHERE ug.user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var groups []*Group

	rows, err := m.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var group Group

		err = rows.Scan(
			&group.ID,
			&group.Name,
		)
		if err != nil {
			return nil, err
		}

		groups = append(groups, &group)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (m GroupModel) InsertGroup(g *Group) error {
	stmt := `INSERT INTO groups
	(name) VALUES ($1)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, g.Name).Scan(&g.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m GroupModel) DeleteGroup(groupID int) error {
	stmt := `DELETE FROM groups
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (m GroupModel) UpdateGroup(g *Group) error {
	stmt := `UPDATE groups
	SET name = $1
	WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, g.Name, g.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m GroupModel) InsertUserIntoGroup(userID, groupID int) error {
	stmt := `INSERT INTO users_groups
	(user_id, group_id)
	VALUES
	($1, $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, userID, groupID)
	if err != nil {
		switch {
		case err.Error() == `ERROR: duplicate key value violates unique constraint "users_groups_pkey" (SQLSTATE 23505)`:
			return ErrUserAlreadyInGroup
		default:
			return err
		}
	}

	return nil
}

func (m GroupModel) RemoveUserFromGroup(userID, groupID int) error {
	stmt := `DELETE FROM users_groups
	WHERE user_id = $1 and group_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}
