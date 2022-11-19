package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNoSuchGroup   = errors.New("no such group")
	ErrNoSuchGroups  = errors.New("no such groups")
	ErrGroupArchived = errors.New("group is archived")
)

type Group struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Archived    bool   `json:"archived"`
	MemberCount *int   `json:"member_count,omitempty"`
}

type GroupModel struct {
	DB *sql.DB
}

func (m GroupModel) GetGroupByID(groupID int) (*Group, error) {
	query := `SELECT id, name, archived
	FROM groups
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var group Group

	err := m.DB.QueryRowContext(ctx, query, groupID).Scan(
		&group.ID,
		&group.Name,
		&group.Archived,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoSuchGroup
		default:
			return nil, err
		}
	}

	return &group, nil
}

func (m GroupModel) GetUserCountForGroup(groupID int) (*int, error) {
	query := `SELECT COUNT (*) FROM users_groups WHERE group_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	err := m.DB.QueryRowContext(ctx, query, groupID).Scan(&count)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

func (m GroupModel) GetAllGroups(archived bool) ([]*Group, error) {
	query := `SELECT g.id, g.name, g.archived, COUNT(ug.user_id)
	FROM groups g
	LEFT JOIN users_groups ug
	ON ug.group_id = g.id
	WHERE g.archived = $1
    GROUP BY g.id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var groups []*Group

	rows, err := m.DB.QueryContext(ctx, query, archived)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var group Group

		err = rows.Scan(
			&group.ID,
			&group.Name,
			&group.Archived,
			&group.MemberCount,
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

func (m GroupModel) GetAllGroupIDsForUser(userID int) ([]int, error) {
	query := `SELECT
	array(
		SELECT g.id
		FROM groups g
		INNER JOIN users_groups ug
		ON ug.group_id = g.id
		WHERE g.archived is FALSE
		AND ug.user_id = $1
	)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ids []int

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(pgtype.NewMap().SQLScanner(&ids))
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m GroupModel) GetUsersByGroupID(groupID int) ([]*User, error) {
	query := `SELECT u.id, u.name, u.role
	FROM users_groups ug
	INNER JOIN users u
	ON ug.user_id = u.id
	WHERE ug.group_id = $1
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, groupID)
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
			&user.Role,
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
	WHERE ug.user_id = $1 AND g.archived is FALSE`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var groups []*Group

	rows, err := m.DB.QueryContext(ctx, query, userID)
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

	err := m.DB.QueryRowContext(ctx, stmt, g.Name).Scan(&g.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m GroupModel) UpdateGroup(g *Group) error {
	stmt := `UPDATE groups
	SET (name, archived) = ($1, $2)
	WHERE id = $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, g.Name, g.Archived, g.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m GroupModel) InsertUserIntoGroup(userID, groupID int) error {
	stmt := `INSERT INTO users_groups
	(user_id, group_id)
	VALUES
	($1, $2)
	ON CONFLICT DO NOTHING`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (m GroupModel) RemoveUserFromGroup(userID, groupID int) error {
	stmt := `DELETE FROM users_groups
	WHERE user_id = $1 and group_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}
