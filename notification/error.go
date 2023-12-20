package notification

import (
	"errors"
	"fmt"
)

type (
	NotificationErrorReason string

	// NotificationError is a custom error for errors occurrs while sending notification to iOS and Android devices
	NotificationError struct {
		err string

		token    string                  // the device token
		deviceOS MobileOS                // the device OS, ither 'ios' or 'android'
		reason   NotificationErrorReason // the reason from the error response given by the push notification services
	}
)

const (
	DefaultResponseReason = "DefaultReason"

	IOSResponseReasonUnregistered   = "Unregistered"
	IOSResponseReasonBadDeviceToken = "BadDeviceToken"

	NotificationErrorReasonIOSUnregistered   = NotificationErrorReason(IOSResponseReasonUnregistered)
	NotificationErrorReasonIOSBadDeviceToken = NotificationErrorReason(IOSResponseReasonBadDeviceToken)

	AndroidResponseReasonMismatchSenderId    = "MismatchSenderId"
	AndroidResponseReasonNotRegistered       = "NotRegistered"
	AndroidResponseReasonInvalidRegistration = "InvalidRegistration"
	AndroidResponseReasonInternalServerError = "InternalServerError"

	NotificationErrorReasonAndroidMismatchSenderId    = NotificationErrorReason(AndroidResponseReasonMismatchSenderId)
	NotificationErrorReasonAndroidNotRegistered       = NotificationErrorReason(AndroidResponseReasonNotRegistered)
	NotificationErrorReasonAndroidInvalidRegistration = NotificationErrorReason(AndroidResponseReasonInvalidRegistration)
	NotificationErrorReasonAndroidInternalServerError = NotificationErrorReason(AndroidResponseReasonInternalServerError)
)

var (
	IOSReason2NotificationErrorReason = map[string]NotificationErrorReason{
		IOSResponseReasonUnregistered:   NotificationErrorReasonIOSUnregistered,
		IOSResponseReasonBadDeviceToken: NotificationErrorReasonIOSBadDeviceToken,
	}

	AndroidReason2NotificationErrorReason = map[string]NotificationErrorReason{
		AndroidResponseReasonMismatchSenderId:    NotificationErrorReasonAndroidMismatchSenderId,
		AndroidResponseReasonNotRegistered:       NotificationErrorReasonAndroidNotRegistered,
		AndroidResponseReasonInvalidRegistration: NotificationErrorReasonAndroidInvalidRegistration,
		AndroidResponseReasonInternalServerError: NotificationErrorReasonAndroidInternalServerError,
	}

	ErrInvalidNotificationErrorReason = errors.New("invalid Notification Error Reason")
	ErrInvalidNotificationDeviceOS    = errors.New("invalid Notification Device OS")
)

func New(err, token string, deviceOS MobileOS, reason NotificationErrorReason) (ne NotificationError) {
	return NotificationError{
		err:      err,
		token:    token,
		deviceOS: deviceOS,
		reason:   reason,
	}
}

func (e *NotificationError) Error() string {
	return fmt.Sprintf("failed sending %s notification to %s: %s - %s", e.deviceOS, e.token, e.reason, e.err)
}

func (e *NotificationError) Token() string {
	return e.token
}

func (e *NotificationError) DeviceOS() MobileOS {
	return e.deviceOS
}

func (e *NotificationError) Reason() NotificationErrorReason {
	return e.reason
}

func GetNotificationErrorReason(deviceOS MobileOS, responseReason string) (NotificationErrorReason, error) {
	var reasonMap map[string]NotificationErrorReason

	switch deviceOS {
	case IOS:
		reasonMap = IOSReason2NotificationErrorReason
	case Android:
		reasonMap = AndroidReason2NotificationErrorReason
	default:
		return "", ErrInvalidNotificationDeviceOS
	}

	reason, ok := reasonMap[responseReason]
	if !ok {
		// There may be other reason values that is not captured in this service,
		// therefore creating them dynamically to record them
		return NotificationErrorReason(responseReason), nil
	}
	return reason, nil
}
