package notification

import (
	"database/sql"
	"time"

	"golang.org/x/net/context"
)

type (
	MobileOS string

	Token struct {
		UserID    string   `json:"user_id"`
		Token     string   `json:"token"`
		DeviceOS  MobileOS `json:"deviceOS"`
		CreatedAt int64    `json:"created_at"`
		ErrorMsg  *string  `json:"error_msg"`
		UpdatedAt int64    `json:"updated_at"`
	}
)

const (
	IOS     = MobileOS("ios")
	Android = MobileOS("android")

	createTokenTable = `CREATE TABLE IF NOT EXISTS notification_token
	(
		user_id    uuid         not null,
		token      varchar(500) not null,
		device_os  varchar(50)  not null,
		created_at bigint       not null,
		error_msg  varchar(500) null default null,
		updated_at bigint       not null default 0,
		PRIMARY KEY (user_id, token, device_os)
	);`

	insertToken = `INSERT INTO notification_token (user_id, token, device_os, created_at, error_msg, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6)
					ON CONFLICT (user_id, token, device_os)
						DO UPDATE SET created_at = $4,
									error_msg  = $5,
									updated_at = $6;`

	findValidTokenByUser = `WITH recent_tokens AS (SELECT user_id, device_os, MAX(created_at) AS created_at
	FROM notification_token
	WHERE user_id = $1
	  AND error_msg IS NULL
	GROUP BY user_id, device_os)
SELECT notification_token.user_id,
notification_token.token,
notification_token.device_os,
notification_token.created_at,
notification_token.error_msg,
notification_token.updated_at
FROM notification_token
INNER JOIN recent_tokens ON notification_token.created_at = recent_tokens.created_at;`

	findAllValidTokens = `WITH recent_tokens AS (SELECT user_id, device_os, MAX(created_at) AS created_at
	FROM notification_token
	WHERE error_msg IS NULL
	GROUP BY user_id, device_os)
SELECT notification_token.user_id,
notification_token.token,
notification_token.device_os,
notification_token.created_at,
notification_token.error_msg,
notification_token.updated_at
FROM notification_token
INNER JOIN recent_tokens ON notification_token.created_at = recent_tokens.created_at;`

	findAllValidUserIDs = `SELECT user_id
	FROM notification_token
	WHERE error_msg IS NULL;`
)

func (t *Token) SetErrorMsg(errMsg *string) {

	if errMsg != nil && *errMsg != "" {
		t.ErrorMsg = errMsg
	} else {
		t.ErrorMsg = nil
	}

	t.UpdatedAt = time.Now().UTC().UnixNano()
}

func (manager *AppMobileNotificationManager) AddNotificationToken(ctx context.Context, token Token) (err error) {

	err = manager.store(ctx, token)

	if err != nil {
		return err
	}

	return nil
}

func (manager *AppMobileNotificationManager) store(ctx context.Context, token Token) (err error) {

	conn := manager.sqlDb.GetDB()

	tx, err := conn.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(insertToken)

	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = stmt.Exec(token.UserID, token.Token, token.DeviceOS, token.CreatedAt, token.ErrorMsg, token.UpdatedAt)

	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil

}

// FindTokens returns the tokens of userID
// or all tokens if passing nil
func (manager *AppMobileNotificationManager) FindTokens(userID *string) (tokens []Token, err error) {

	var rows *sql.Rows

	if userID == nil || *userID == "" {
		rows, err = manager.sqlDb.GetDB().Query(findAllValidTokens)
	} else {
		rows, err = manager.sqlDb.GetDB().Query(findValidTokenByUser, *userID)
	}

	if err != nil {
		return tokens, err
	}

	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {

		var token Token

		err = rows.Scan(&token.UserID, &token.Token, &token.DeviceOS, &token.CreatedAt, &token.ErrorMsg, &token.UpdatedAt)

		if err != nil {
			return tokens, err
		}

		tokens = append(tokens, token)

	}

	return tokens, err
}

// FindUserIDs returns the userIDs of all valid tokens
func (manager *AppMobileNotificationManager) FindUserIDs() (userIDs []string, err error) {

	var rows *sql.Rows

	rows, err = manager.sqlDb.GetDB().Query(findAllValidUserIDs)

	if err != nil {
		return userIDs, err
	}

	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {

		var userID string

		err = rows.Scan(&userID)

		if err != nil {
			return userIDs, err
		}

		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}
