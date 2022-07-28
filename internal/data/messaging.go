package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/microcosm-cc/bluemonday"
)

var (
	ErrUserAlreadyInThread    = errors.New("user already in thread")
	ErrUserNotInThread        = errors.New("user not in thread")
	ErrUsersNotInThread       = errors.New("users not in thread")
	ErrNoSuchThread           = errors.New("no such thread")
	ErrNoSuchMessage          = errors.New("no such message")
	ErrThreadAlreadyLocked    = errors.New("thread already locked")
	ErrThreadAlreadyUnlocked  = errors.New("thread already unlocked")
	ErrCantDeleteFirstMessage = errors.New("can't delete first message of thread")
)

const (
	ActionAddedUser   = "added_user"
	ActionRemovedUser = "removed_user"
	ActionLocked      = "locked"
	ActionUnlocked    = "unlocked"
)

const (
	MsgTypeNormal      = "normal"
	MsgTypeThreadStart = "thread_start"
)

type Thread struct {
	ID           int       `json:"id"`
	User         *User     `json:"user"`
	Title        string    `json:"title"`
	Locked       bool      `json:"locked"`
	MessageCount *int      `json:"message_count,omitempty"`
	Read         *bool     `json:"read,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Message struct {
	ID        int       `json:"id"`
	ThreadID  int       `json:"thread_id"`
	User      *User     `json:"user"`
	Body      string    `json:"body,omitempty"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"-"`
}

type MessagingModel struct {
	DB        *pgxpool.Pool
	XSSpolicy *bluemonday.Policy
}

func (m MessagingModel) GetThreadByID(threadID int) (*Thread, error) {
	query := `SELECT t.id, t.user_id, u.name, u.role, t.title, t.locked, t.created_at, t.updated_at
	FROM threads t
	INNER JOIN users u
	ON t.user_id = u.id
	WHERE t.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var thread Thread
	thread.User = &User{}

	err := m.DB.QueryRow(ctx, query, threadID).Scan(
		&thread.ID,
		&thread.User.ID,
		&thread.User.Name,
		&thread.User.Role,
		&thread.Title,
		&thread.Locked,
		&thread.CreatedAt,
		&thread.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchThread
		default:
			return nil, err
		}
	}

	return &thread, nil
}

func (m MessagingModel) InsertThread(t *Thread) error {
	stmt := `INSERT INTO threads
	(user_id, title, locked, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, t.User.ID, t.Title, t.Locked, t.CreatedAt, t.UpdatedAt).Scan(&t.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m MessagingModel) DeleteThread(threadID int) error {
	stmt := `DELETE FROM threads
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, threadID)
	if err != nil {
		return err
	}
	return nil
}

func (m MessagingModel) AddUserToThread(threadID, userID int) error {
	stmt := `INSERT INTO threads_recipients
	(thread_id, user_id)
	VALUES
	($1, $2)
	ON CONFLICT ON CONSTRAINT threads_recipients_pkey
	DO UPDATE SET group_id = NULL`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, threadID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) RemoveUserFromThread(threadID, userID int) error {
	stmt := `DELETE FROM threads_recipients
	WHERE thread_id = $1 and user_id = $2 and group_id is NULL`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, threadID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) AddGroupToThread(threadID, groupID int) error {
	stmt := `INSERT INTO threads_recipients (thread_id, user_id, group_id)
	( SELECT $1, ug.user_id, ug.group_id
	FROM users_groups ug
	WHERE ug.group_id = $2 )
	ON CONFLICT ON CONSTRAINT threads_recipients_pkey
	DO UPDATE SET group_id = excluded.group_id
    WHERE threads_recipients.group_id IS NOT NULL`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, threadID, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) RemoveGroupFromThread(threadID, groupID int) error {
	stmt := `DELETE FROM threads_recipients
	WHERE thread_id = $1 and group_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, threadID, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) GetUsersInThread(threadID int) ([]*User, error) {
	query := `SELECT tr.user_id, u.name, u.role
	FROM threads_recipients tr
	INNER JOIN users u
	ON tr.user_id = u.id
	WHERE tr.thread_id = $1 and tr.group_id is NULL
	ORDER BY u.name ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, threadID)
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

func (m MessagingModel) GetGroupsInThread(threadID int) ([]*Group, error) {
	query := `SELECT DISTINCT tr.group_id, g.name
	FROM threads_recipients tr
	INNER JOIN groups g
	ON tr.group_id = g.id
	WHERE tr.thread_id = $1 and tr.group_id is NOT NULL
	ORDER BY g.name ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, threadID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var groups []*Group

	for rows.Next() {
		var group Group
		err := rows.Scan(
			&group.ID,
			&group.Name,
		)
		if err != nil {
			return nil, err
		}
		groups = append(groups, &group)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (m MessagingModel) GetMessageByID(messageID int) (*Message, error) {
	query := `SELECT m.id, m.thread_id, m.user_id, u.name, u.role, m.body, m.type, m.created_at, m.updated_at, m.version
	FROM messages m
	INNER JOIN users u
	ON m.user_id = u.id
	WHERE m.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var message Message
	message.User = &User{}

	err := m.DB.QueryRow(ctx, query, messageID).Scan(
		&message.ID,
		&message.ThreadID,
		&message.User.ID,
		&message.User.Name,
		&message.User.Role,
		&message.Body,
		&message.Type,
		&message.CreatedAt,
		&message.UpdatedAt,
		&message.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchMessage
		default:
			return nil, err
		}
	}

	return &message, nil
}

func (m MessagingModel) InsertMessage(ms *Message) error {
	stmt := `INSERT INTO messages
	(thread_id, user_id, body, type, created_at, updated_at, version)
	VALUES
	($1, $2, $3, $4, $5, $6, $7)
	RETURNING id`

	sanitizedHTML := m.XSSpolicy.Sanitize(ms.Body)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, ms.ThreadID, ms.User.ID, sanitizedHTML, ms.Type, ms.CreatedAt, ms.UpdatedAt, ms.Version).Scan(&ms.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m MessagingModel) DeleteMessage(messageID int) error {
	stmt := `DELETE FROM messages
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, messageID)
	if err != nil {
		return err
	}
	return nil
}

func (m MessagingModel) UpdateMessage(ms *Message) error {
	stmt := `UPDATE messages
	SET (body, updated_at, version)
	= ($1, $2, version+1)
	WHERE id = $3 and version = $4
	RETURNING version`

	sanitizedHTML := m.XSSpolicy.Sanitize(ms.Body)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, sanitizedHTML, ms.UpdatedAt, ms.ID, ms.Version).Scan(&ms.Version)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MessagingModel) GetAllMessagesByThreadID(threadID int) ([]*Message, error) {
	query := `SELECT m.id, m.thread_id, m.user_id, u.name, u.role, m.body, m.type, m.created_at, m.updated_at, m.version
	FROM messages m
	INNER JOIN users u
	ON m.user_id = u.id
	WHERE thread_id = $1
	ORDER BY created_at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, threadID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var messages []*Message

	for rows.Next() {
		var message Message
		message.User = &User{}

		err = rows.Scan(
			&message.ID,
			&message.ThreadID,
			&message.User.ID,
			&message.User.Name,
			&message.User.Role,
			&message.Body,
			&message.Type,
			&message.CreatedAt,
			&message.UpdatedAt,
			&message.Version,
		)
		if err != nil {
			return nil, err
		}

		messages = append(messages, &message)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (m MessagingModel) GetThreadsForUser(userID int) ([]*Thread, error) {
	query := `SELECT t.id, t.user_id, u.name, u.role, t.title, t.locked, tr.read, t.created_at, t.updated_at
	FROM threads t
	INNER JOIN threads_recipients tr
	ON t.id = tr.thread_id
	INNER JOIN users u
	on t.user_id = u.id
	WHERE tr.user_id = $1
	ORDER BY updated_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var threads []*Thread

	for rows.Next() {
		var thread Thread
		thread.User = &User{}

		err = rows.Scan(
			&thread.ID,
			&thread.User.ID,
			&thread.User.Name,
			&thread.User.Role,
			&thread.Title,
			&thread.Locked,
			&thread.Read,
			&thread.CreatedAt,
			&thread.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		threads = append(threads, &thread)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return threads, nil
}

func (m MessagingModel) IsUserInThread(userID, threadID int) (bool, error) {
	query := `SELECT COUNT(1) FROM threads_recipients
	WHERE thread_id = $1 and user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRow(ctx, query, threadID, userID).Scan(&result)
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (m MessagingModel) GetMessageCountForThread(threadID int) (int, error) {
	query := `SELECT COUNT (*) FROM messages WHERE thread_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	err := m.DB.QueryRow(ctx, query, threadID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m MessagingModel) SetThreadAsReadForUser(threadID, userID int) error {
	stmt := `UPDATE threads_recipients
	SET read = true
	WHERE thread_id = $1 and user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, threadID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) SetThreadAsUnreadForAll(threadID int) error {
	stmt := `UPDATE threads_recipients
	SET read = false
	WHERE thread_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, threadID)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) SetThreadLocked(threadID int, locked bool) error {
	stmt := `UPDATE threads
	SET locked = $1
	WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, locked, threadID)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) SetThreadUpdatedAt(threadID int) error {
	stmt := `UPDATE threads
	SET updated_at = $1
	WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, time.Now().UTC(), threadID)
	if err != nil {
		return err
	}

	return nil
}
