package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrUserAlreadyInThread   = errors.New("user already in thread")
	ErrUserNotInThread       = errors.New("user not in thread")
	ErrUsersNotInThread      = errors.New("users not in thread")
	ErrNoSuchThread          = errors.New("no such thread")
	ErrNoSuchMessage         = errors.New("no such message")
	ErrThreadAlreadyLocked   = errors.New("thread already locked")
	ErrThreadAlreadyUnlocked = errors.New("thread already unlocked")
)

const (
	ActionAddedUser   = "added_user"
	ActionRemovedUser = "removed_user"
	ActionLocked      = "locked"
	ActionUnlocked    = "unlocked"
)

type Thread struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Locked    bool      `json:"locked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID        int       `json:"id"`
	ThreadID  int       `json:"thread_id"`
	UserID    int       `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"-"`
}

type ThreadLog struct {
	Action  string    `json:"action"`
	Targets []int     `json:"target"`
	By      int       `json:"by"`
	At      time.Time `json:"at"`
}

type MessagingModel struct {
	DB *pgxpool.Pool
}

func (m MessagingModel) GetThreadByID(threadID int) (*Thread, error) {
	query := `SELECT id, user_id, title, body, locked, created_at, updated_at
	FROM threads
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var thread Thread

	err := m.DB.QueryRow(ctx, query, threadID).Scan(
		&thread.ID,
		&thread.UserID,
		&thread.Title,
		&thread.Body,
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
	(user_id, title, body, locked, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, t.UserID, t.Title, t.Body, t.Locked, t.CreatedAt, t.UpdatedAt).Scan(&t.ID)
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

func (m MessagingModel) UpdateThread(t *Thread) error {
	stmt := `UPDATE threads
	SET (title, body, locked, updated_at)
	= ($1, $2, $3, $4)
	WHERE id = $5`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, t.Title, t.Body, t.Locked, t.UpdatedAt, t.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m MessagingModel) AddUserToThread(userID, threadID int) error {
	stmt := `INSERT INTO users_threads
	(user_id, thread_id)
	VALUES
	($1, $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, userID, threadID)
	if err != nil {
		switch {
		case err.Error() == `ERROR: duplicate key value violates unique constraint "users_threads_pkey" (SQLSTATE 23505)`:
			return ErrUserAlreadyInThread
		default:
			return err
		}
	}

	return nil
}

func (m MessagingModel) RemoveUserFromThread(userID, threadID int) error {
	stmt := `DELETE FROM users_threads
	WHERE user_id = $1 and thread_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, userID, threadID)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) InsertThreadLog(tl *ThreadLog) error {
	stmt := `INSERT INTO thread_log
	(action, target, by, at)
	VALUES
	($1, $2, $3, $4)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, tl.Action, tl.Targets, tl.By, tl.At)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) GetMessageByID(messageID int) (*Message, error) {
	query := `SELECT id, thread_id, user_id, body, created_at, updated_at, version
	FROM messages
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var message Message

	err := m.DB.QueryRow(ctx, query, messageID).Scan(
		&message.ID,
		&message.ThreadID,
		&message.UserID,
		&message.Body,
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
	(thread_id, user_id, body, created_at, updated_at, version)
	VALUES
	($1, $2, $3, $4, $5, $6)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, ms.ThreadID, ms.UserID, ms.Body, ms.CreatedAt, ms.UpdatedAt, ms.Version).Scan(&ms.ID)
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, ms.Body, ms.UpdatedAt, ms.ID, ms.Version).Scan(&ms.Version)
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
	query := `SELECT id, thread_id, user_id, body, created_at, updated_at, version
	FROM messages
	WHERE thread_id = $1`

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

		err = rows.Scan(
			&message.ID,
			&message.ThreadID,
			&message.UserID,
			&message.Body,
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
	query := `SELECT t.id, t.user_id, t.title, t.body, t.locked, t.created_at, t.updated_at
	FROM threads t
	INNER JOIN users_threads ut
	ON t.id = ut.thread_id
	WHERE ut.user_id = $1`

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

		err = rows.Scan(
			&thread.ID,
			&thread.UserID,
			&thread.Title,
			&thread.Body,
			&thread.Locked,
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

func (m MessagingModel) GetThreadIDsForUser(userID int) ([]int, error) {
	query := `SELECT
	array(SELECT thread_id
		FROM users_threads
		WHERE user_id = $1)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ids []int

	err := m.DB.QueryRow(ctx, query, userID).Scan(&ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m MessagingModel) GetUserIDsForThread(threadID int) ([]int, error) {
	query := `SELECT
	array(SELECT user_id
		FROM users_threads
		WHERE thread_id = $1)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ids []int

	err := m.DB.QueryRow(ctx, query, threadID).Scan(&ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}
