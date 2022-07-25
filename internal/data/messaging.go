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
	ID    int    `json:"id"`
	User  *User  `json:"user"`
	Title string `json:"title"`
	// Body         string    `json:"body"`
	Locked       bool      `json:"locked"`
	MessageCount int       `json:"message_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Message struct {
	ID        int       `json:"id"`
	ThreadID  int       `json:"thread_id"`
	User      *User     `json:"user"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"-"`
}

// type ThreadLog struct {
// 	ThreadID int    `json:"thread_id"`
// 	Action   string `json:"action"`
// 	// Targets  []int     `json:"target"`
// 	Targets []*User   `json:"targets"`
// 	By      int       `json:"by"`
// 	At      time.Time `json:"at"`
// }

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

func (m MessagingModel) UpdateThread(t *Thread) error {
	stmt := `UPDATE threads
	SET (title, locked, updated_at)
	= ($1, $2, $3)
	WHERE id = $4`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, t.Title, t.Locked, t.UpdatedAt, t.ID)
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

// func (m MessagingModel) InsertThreadLog(tl *ThreadLog) error {
// 	stmt := `INSERT INTO thread_log
// 	(thread_id, action, target, by, at)
// 	VALUES
// 	($1, $2, $3, $4, $5)`

// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	var targets []int

// 	for _, u := range tl.Targets {
// 		targets = append(targets, u.ID)
// 	}

// 	_, err := m.DB.Exec(ctx, stmt, tl.ThreadID, tl.Action, targets, tl.By, tl.At)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (m MessagingModel) GetMessageByID(messageID int) (*Message, error) {
	query := `SELECT m.id, m.thread_id, m.user_id, u.name, u.role, m.body, m.created_at, m.updated_at, m.version
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

	sanitizedHTML := m.XSSpolicy.Sanitize(ms.Body)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, ms.ThreadID, ms.User.ID, sanitizedHTML, ms.CreatedAt, ms.UpdatedAt, ms.Version).Scan(&ms.ID)
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
	query := `SELECT m.id, m.thread_id, m.user_id, u.name, u.role, m.body, m.created_at, m.updated_at, m.version
	FROM messages m
	INNER JOIN users u
	ON m.user_id = u.id
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
		message.User = &User{}

		err = rows.Scan(
			&message.ID,
			&message.ThreadID,
			&message.User.ID,
			&message.User.Name,
			&message.User.Role,
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

// func (m MessagingModel) GetAllLogsByThreadID(threadID int) ([]*ThreadLog, error) {
// 	query := `SELECT thread_id, action, target, by, at
// 	FROM thread_log
// 	WHERE thread_id = $1`

// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	rows, err := m.DB.Query(ctx, query, threadID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer rows.Close()

// 	var threadLogs []*ThreadLog

// 	for rows.Next() {
// 		var tl ThreadLog
// 		var targets []int

// 		err = rows.Scan(
// 			&tl.ThreadID,
// 			&tl.Action,
// 			&targets,
// 			&tl.By,
// 			&tl.At,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}

// 		for _, id := range targets {
// 			tl.Targets = append(tl.Targets, &User{ID: id})
// 		}

// 		threadLogs = append(threadLogs, &tl)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return threadLogs, nil
// }

func (m MessagingModel) GetThreadsForUser(userID int) ([]*Thread, error) {
	query := `SELECT t.id, t.user_id, u.name, u.role, t.title, t.locked, t.created_at, t.updated_at
	FROM threads t
	INNER JOIN users_threads ut
	ON t.id = ut.thread_id
	INNER JOIN users u
	on t.user_id = u.id
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
		thread.User = &User{}

		err = rows.Scan(
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

func (m MessagingModel) GetUsersForThread(threadID int) ([]*User, error) {
	query := `SELECT id, name, role
	FROM users u
	INNER JOIN users_threads ut
	ON u.id = ut.user_id
	WHERE ut.thread_id = $1`

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

func (m MessagingModel) IsUserInThread(userID, threadID int) (bool, error) {
	query := `SELECT COUNT(1) FROM users_threads
	WHERE user_id = $1 and thread_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRow(ctx, query, userID, threadID).Scan(&result)
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
