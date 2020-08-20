package notification

import (
	"github.com/HelloSundayMorning/apputils/app"
	"github.com/HelloSundayMorning/apputils/db"
	"github.com/HelloSundayMorning/apputils/log"
	"github.com/Smarp/fcm-http"
	apns "github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"golang.org/x/net/context"
	"time"
)

type (
	MobileOS string

	Token struct {
		UserID    string
		Token     string
		DeviceOS  MobileOS
		CreatedAt int64
		UpdatedAt int64
	}

	MobileNotificationManager interface {
		AddNotificationToken(ctx context.Context, userID string, token string, deviceOs MobileOS) (err error)
		SendAlert(ctx context.Context, userID, title, message string) (err error)
	}

	AppMobileNotificationManager struct {
		AppID     app.ApplicationID
		sqlDb     db.AppSqlDb
		IosClient *apns.Client
		FcmClient *fcm.Sender
	}
)

const (
	IOS     = MobileOS("ios")
	Android = MobileOS("android")

	component = "mobileNotificationManager"

	createTokenTable = `CREATE TABLE IF NOT EXISTS notification_token (
									 user_id                    uuid                      not null,
									 token                      varchar(500)              not null,
									 device_os                  varchar(50)               not null,
                                     created_at                 bigint                    not null,
                                     PRIMARY KEY (user_id, token, device_os));`

	insertToken = `INSERT INTO notification_token (user_id, token, device_os, created_at , updated_at)
                                VALUES ($1, $2, $3, $4, $5)
                                ON CONFLICT (user_id, token, device_os) DO UPDATE SET
                                updated_at = $5`

	findToken = `SELECT user_id, token, device_os, created_at, updated_at
                    FROM notification_token
                    WHERE user_id = $1`
)

func NewMobileNotificationManagerApp(appID app.ApplicationID, sqlDb db.AppSqlDb, iOsClient *apns.Client, fcmClient *fcm.Sender) (manager *AppMobileNotificationManager, err error) {

	manager = &AppMobileNotificationManager{
		AppID:     appID,
		sqlDb:     sqlDb,
		IosClient: iOsClient,
		FcmClient: fcmClient,
	}

	err = manager.initialize()

	if err != nil {
		return nil, err
	}

	log.PrintfNoContext(manager.AppID, component, "MobileNotificationManager initialized")

	return manager, nil

}

func (manager *AppMobileNotificationManager) initialize() (err error) {

	conn := manager.sqlDb.GetDB()

	ctx := context.Background()

	tx, err := conn.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(createTokenTable)

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = stmt.Exec()

	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}

func (manager *AppMobileNotificationManager) AddNotificationToken(ctx context.Context, userID string, token string, deviceOs MobileOS) (err error) {

	err = manager.store(ctx, Token{
		UserID:    userID,
		Token:     token,
		DeviceOS:  deviceOs,
		CreatedAt: time.Now().UTC().UnixNano(),
		UpdatedAt: time.Now().UTC().UnixNano(),
	})

	if err != nil {
		return err
	}

	return nil
}

func (manager *AppMobileNotificationManager) store(ctx context.Context, token Token) (err error) {

	conn := manager.sqlDb.GetDB()

	tx, err := conn.BeginTx(ctx, nil)

	stmt, err := tx.Prepare(insertToken)

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = stmt.Exec(token.UserID, token.Token, token.DeviceOS, token.CreatedAt, token.UpdatedAt)

	if err != nil {
		tx.Rollback()
		return err
	}

	return nil

}

func (manager *AppMobileNotificationManager) findTokensByUser(userID string) (tokens []Token, err error) {

	rows, err := manager.sqlDb.GetDB().Query(findToken, userID)

	if err != nil {
		return tokens, err
	}

	defer rows.Close()

	if rows.Next() {

		var token Token

		err = rows.Scan(&token.UserID, &token.Token, &token.DeviceOS, &token.CreatedAt, &token.UpdatedAt)

		if err != nil {
			return tokens, err
		}

		tokens = append(tokens, token)

	}

	return tokens, err
}

func (manager *AppMobileNotificationManager) SendAlert(ctx context.Context, userID, title, message string) (err error) {

	tokens, err := manager.findTokensByUser(userID)

	for _, token := range tokens {

		switch token.DeviceOS {
		case IOS:
			err = manager.sendIOSAlert(ctx, token.Token, title, message)

			if err != nil {
				return err
			}

		case Android:
			err = manager.sendAndroidAlert(ctx, token.Token, title, message)

			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (manager *AppMobileNotificationManager) SendDataNotification(ctx context.Context, userID, title, message string, customData map[string]interface{}) (err error) {

	tokens, err := manager.findTokensByUser(userID)

	for _, token := range tokens {

		switch token.DeviceOS {
		case IOS:
			err = manager.sendIOSDataNotification(ctx, token.Token, title, message, customData)

			if err != nil {
				return err
			}

		case Android:
			err = manager.sendAndroidDataNotification(ctx, token.Token, title, message, customData)

			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (manager *AppMobileNotificationManager) sendIOSDataNotification(ctx context.Context, userDeviceToken, title, message string, customData map[string]interface{}) (err error) {

	dataPayload := payload.NewPayload().
		Alert(title).
		Badge(0).
		Sound("default")

	for k, v := range customData {
		dataPayload.Custom(k, v)
	}

	notification := apns.Notification{
		Topic: "io.daybreakapp.app",
		Payload: dataPayload,
		DeviceToken: userDeviceToken,
	}

	r, err := manager.IosClient.Push(&notification)

	if err != nil {
		return err
	}

	if r.Sent() {
		log.Printf(ctx, component, "Sent IOS Data notification, Status %d, ID %s", r.StatusCode, r.ApnsID)
	} else {
		log.Printf(ctx, component, "Fail to send IOS Data notification, Status %d, ID %s, Reason %s", r.StatusCode, r.ApnsID, r.Reason)
	}

	return nil

}

func (manager *AppMobileNotificationManager) sendIOSAlert(ctx context.Context, userDeviceToken, title, message string) (err error) {

	notification := apns.Notification{
		Topic: "io.daybreakapp.app",
		Payload: payload.NewPayload().
			Alert(title).
			Badge(0).
			Sound("default").
			AlertBody(message),
		DeviceToken: userDeviceToken,
	}

	r, err := manager.IosClient.Push(&notification)

	if err != nil {
		return err
	}

	if r.Sent() {
		log.Printf(ctx, component, "Sent IOS Alert, Status %d, ID %s", r.StatusCode, r.ApnsID)
	} else {
		log.Printf(ctx, component, "Fail to send IOS Alert, Status %d, ID %s, Reason %s", r.StatusCode, r.ApnsID, r.Reason)
	}

	return nil
}

func (manager *AppMobileNotificationManager) sendAndroidDataNotification(ctx context.Context, userDeviceToken, title, message string, customData map[string]interface{}) (err error) {

	regIDs := []string{userDeviceToken}
	msg := &fcm.Message{
		RegistrationIDs:       regIDs,
		CollapseKey:           "",
		Data:                  customData,
		DelayWhileIdle:        false,
		TimeToLive:            0,
		RestrictedPackageName: "",
		DryRun:                false,
		Notification: fcm.Notification{
			Title: title,
			Body:  message,
		},
	}

	r, err := manager.FcmClient.Send(msg, 2)
	if err != nil {
		return err
	}

	if r.Success == 1 {
		log.Printf(ctx, component, "Sent Android Data Notification")
	} else {
		log.Printf(ctx, component, "Fail to send Data Notification")
	}

	return nil
}

func (manager *AppMobileNotificationManager) sendAndroidAlert(ctx context.Context, userDeviceToken, title, message string) (err error) {

	regIDs := []string{userDeviceToken}
	msg := &fcm.Message{
		RegistrationIDs:       regIDs,
		CollapseKey:           "",
		Data:                  nil,
		DelayWhileIdle:        false,
		TimeToLive:            0,
		RestrictedPackageName: "",
		DryRun:                false,
		Notification: fcm.Notification{
			Title: title,
			Body:  message,
		},
	}

	r, err := manager.FcmClient.Send(msg, 2)
	if err != nil {
		return err
	}

	if r.Success == 1 {
		log.Printf(ctx, component, "Sent Android Alert")
	} else {
		log.Printf(ctx, component, "Fail to send Android Alert")
	}

	return nil
}
