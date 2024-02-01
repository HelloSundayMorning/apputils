package notification

import (
	"errors"
	"fmt"
	"time"

	"github.com/HelloSundayMorning/apputils/app"
	"github.com/HelloSundayMorning/apputils/db"
	"github.com/HelloSundayMorning/apputils/log"
	"github.com/Smarp/fcm-http"
	apns "github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"golang.org/x/net/context"
)

type (
	MobileNotificationManager interface {
		AddNotificationToken(ctx context.Context, userID string, token string, deviceOs MobileOS) (err error)
		SendAlert(ctx context.Context, userID, title, message string) (err error)
		FindUserIDs() (userIDs []string, err error)
	}

	AppMobileNotificationManager struct {
		AppID     app.ApplicationID
		sqlDb     db.AppSqlDb
		IosClient *apns.Client
		FcmClient *fcm.Sender
	}
)

const (
	component = "mobileNotificationManager"
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
		_ = tx.Rollback()
		return err
	}

	_, err = stmt.Exec()

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

func (manager *AppMobileNotificationManager) SendAlert(ctx context.Context, userID, title, message string) (err error) {

	tokens, err := manager.findTokens(&userID)

	if err != nil {
		log.Errorf(ctx, component, "Error finding tokens for user %s", userID)
		return err
	}

	log.Printf(ctx, component, "Found %d tokens for user %s. Trying to send all", len(tokens), userID)

	for _, token := range tokens {

		log.Printf(ctx, component, "Trying %s token created %v", token.DeviceOS, time.Unix(0, token.CreatedAt))

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

// BroadcastDataNotification sends a push notification to the given userID
// On error, it returns an instance of DataNotificationError
func (manager *AppMobileNotificationManager) SendDataNotification(ctx context.Context, userID, title, message string, customData map[string]interface{}) (err error) {

	tokens, err := manager.findTokens(&userID)

	if err != nil {
		log.Errorf(ctx, component, "Error finding tokens for user %s", userID)
		return err
	}

	log.Printf(ctx, component, "Found %d tokens for user %s. Trying to send all", len(tokens), userID)

	return manager.sendDataNotifications(ctx, tokens, title, message, customData)
}

// BroadcastDataNotification sends a push notification to all available tokens
// On error, it returns an instance of DataNotificationError
func (manager *AppMobileNotificationManager) BroadcastDataNotification(ctx context.Context, title, message string, customData map[string]interface{}) (err error) {

	tokens, err := manager.findTokens(nil)

	if err != nil {
		log.Errorf(ctx, component, "Error finding tokens for broadcasting (%s)", err)
		return err
	}

	log.Printf(ctx, component, "Found %d tokens for broadcasting. Trying to send all", len(tokens))

	return manager.sendDataNotifications(ctx, tokens, title, message, customData)
}

func (manager *AppMobileNotificationManager) sendDataNotifications(ctx context.Context, tokens []Token, title, message string, customData map[string]interface{}) (err error) {

	// notificationErrors is a collection of (Token, NotificationRequestError) pairs
	var notificationErrors []NotificationError

	for _, token := range tokens {

		err = manager.sendDataNotification(ctx, token, title, message, customData)

		var notificationRequestError NotificationRequestError

		// accumulate only notificationRequestErrors
		// log other errors and let push notifications continue being sent
		if err != nil {
			if errors.As(err, &notificationRequestError) {
				notificationErrors = append(notificationErrors, NotificationError{Token: token, Err: notificationRequestError})
			} else {
				log.Printf(ctx, component, "Failed to send notification to token %#v: %s", token, err)
			}
		}
	}

	if len(notificationErrors) > 0 {
		// Users might delete app and reinstall => that'll create new token and the existed one will be invalid.
		log.Printf(ctx, component, "Failed to send notification to tokens %#v\n", notificationErrors)

		return &DataNotificationError{
			NotificationErrors: notificationErrors,
		}
	}

	return nil
}

// SendDataNotificationByToken sends a single push notification by the token param
// On error, it returns an wrapped NotificationRequestError instance.
func (manager *AppMobileNotificationManager) SendDataNotificationByToken(ctx context.Context, token Token, title, message string, customData map[string]interface{}) (err error) {

	return manager.sendDataNotification(ctx, token, title, message, customData)
}

func (manager *AppMobileNotificationManager) sendDataNotification(ctx context.Context, token Token, title, message string, customData map[string]interface{}) (err error) {

	log.Printf(ctx, component, "Trying %s token, created at %v", token.DeviceOS, time.Unix(0, token.CreatedAt))

	switch token.DeviceOS {
	case IOS:
		err = manager.sendIOSDataNotification(ctx, token.Token, title, message, customData)

	case Android:
		err = manager.sendAndroidDataNotification(ctx, token.Token, title, message, customData)

	default:
		err = fmt.Errorf("invalid DeviceOS: %s", token.DeviceOS)
	}

	if err != nil {
		return err
	}
	return nil
}

// sendIOSDataNotification sends a single iOS notification
//
//	returns either a NotificationRequestError or a generic error
func (manager *AppMobileNotificationManager) sendIOSDataNotification(ctx context.Context, userDeviceToken, title, message string, customData map[string]interface{}) (err error) {

	dataPayload := payload.NewPayload().
		AlertTitle(title).
		AlertBody(message).
		Badge(0).
		Sound("default")

	for k, v := range customData {
		dataPayload.Custom(k, v)
	}

	notification := apns.Notification{
		Topic:       "io.daybreakapp.app",
		Payload:     dataPayload,
		DeviceToken: userDeviceToken,
	}

	r, err := manager.IosClient.Push(&notification)

	if err != nil {
		log.Printf(ctx, component, "Error sending IOS Data notification. Error: %s", err)
		return err
	}

	if !r.Sent() {
		log.Printf(ctx, component, "Fail to send IOS Data notification, Status %d, ID %s, Reason %s", r.StatusCode, r.ApnsID, r.Reason)

		reason, err := GetNotificationRequestErrorReason(IOS, r.Reason)
		if err != nil {
			log.Printf(ctx, component, "Fail to get IOS Data notification error reason: %s", err)

			reason = DefaultNotificationResponseReason

			// no need to reture here, the purpose is to capture the error state
		}

		err = NotificationRequestError{
			ErrMsg:   "fail to send IOS Data notification",
			TokenStr: userDeviceToken,
			DeviceOS: IOS,
			Reason:   reason,
		}

		return err
	}

	log.Printf(ctx, component, "Sent IOS Data notification, Status %d, ID %s", r.StatusCode, r.ApnsID)

	return nil

}

func (manager *AppMobileNotificationManager) sendIOSAlert(ctx context.Context, userDeviceToken, title, message string) (err error) {

	notification := apns.Notification{
		Topic: "io.daybreakapp.app",
		Payload: payload.NewPayload().
			AlertTitle(title).
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

// sendAndroidDataNotification sends a single Android notification
//
//	returns either a NotificationRequestError or a generic error
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
		log.Printf(ctx, component, "Error sending Android Data notification. Error: %s", err)
		return err
	}

	if r.Success != 1 {
		log.Printf(ctx, component, "Fail to send Android Data Notification. Failure code %d, Results %v, id %d", r.Failure, r.Results, r.CanonicalIDs)

		var reason NotificationRequestErrorReason

		if len(r.Results) > 0 {
			// r.Results is in the form of [{ MismatchSenderId}]
			reason, err = GetNotificationRequestErrorReason(Android, r.Results[0].Error)
			if err != nil {
				log.Printf(ctx, component, "Fail to get Android Data notification error reason: %s", err)
				reason = DefaultNotificationResponseReason

				// no need to reture here, the purpose is to capture the error state
			}
		} else {
			log.Printf(ctx, component, "Fail to get GCM response results: %s", r.Results)

			reason = DefaultNotificationResponseReason
		}

		err = NotificationRequestError{
			ErrMsg:   "fail to send Android Data notification",
			TokenStr: userDeviceToken,
			DeviceOS: IOS,
			Reason:   reason,
		}

		return err
	}

	log.Printf(ctx, component, "Sent Android Data Notification. Results %v, id %d", r.Results, r.CanonicalIDs)

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
		log.Printf(ctx, component, "Sent Android Alert. Results %v, id %d", r.Results, r.CanonicalIDs)
	} else {
		log.Printf(ctx, component, "Fail to send Android Alert. Failure code %d, Results %v, id %d", r.Failure, r.Results, r.CanonicalIDs)
	}

	return nil
}
