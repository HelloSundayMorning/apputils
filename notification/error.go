package notification

import (
	"errors"
	"fmt"
)

type (
	// DataNotificationError represents all errors from calling `sendDataNotification()`
	// It holds a slice of NotificationError instances
	DataNotificationError struct {
		NotificationErrors []NotificationError
	}

	// NotificationError represents a single notification error
	// It holds an instance of the underlying error, possible an instance of NotificationRequestError
	NotificationError struct {
		Token Token
		Err   error
	}

	// NotificationRequestErrorReason represents the error categories set by the iOS and the Android push notification service
	NotificationRequestErrorReason string

	// NotificationRequestError is a custom error for errors occurrs while sending notification to iOS and Android devices
	NotificationRequestError struct {
		ErrMsg string

		TokenStr string                         // the device token
		DeviceOS MobileOS                       // the device OS, ither 'ios' or 'android'
		Reason   NotificationRequestErrorReason // the reason from the error response given by the push notification services
	}
)

const (
	DefaultNotificationResponseReason = "DefaultReason"

	IOSNotificationResponseReasonUnregistered   = "Unregistered"
	IOSNotificationResponseReasonBadDeviceToken = "BadDeviceToken"

	NotificationRequestErrorReasonIOSUnregistered   = NotificationRequestErrorReason(IOSNotificationResponseReasonUnregistered)
	NotificationRequestErrorReasonIOSBadDeviceToken = NotificationRequestErrorReason(IOSNotificationResponseReasonBadDeviceToken)

	AndroidNotificationResponseReasonMismatchSenderId    = "MismatchSenderId"
	AndroidNotificationResponseReasonNotRegistered       = "NotRegistered"
	AndroidNotificationResponseReasonInvalidRegistration = "InvalidRegistration"
	AndroidNotificationResponseReasonInternalServerError = "InternalServerError"

	NotificationRequestErrorReasonAndroidMismatchSenderId    = NotificationRequestErrorReason(AndroidNotificationResponseReasonMismatchSenderId)
	NotificationRequestErrorReasonAndroidNotRegistered       = NotificationRequestErrorReason(AndroidNotificationResponseReasonNotRegistered)
	NotificationRequestErrorReasonAndroidInvalidRegistration = NotificationRequestErrorReason(AndroidNotificationResponseReasonInvalidRegistration)
	NotificationRequestErrorReasonAndroidInternalServerError = NotificationRequestErrorReason(AndroidNotificationResponseReasonInternalServerError)
)

var (
	IOSReason2NotificationErrorReason = map[string]NotificationRequestErrorReason{
		IOSNotificationResponseReasonUnregistered:   NotificationRequestErrorReasonIOSUnregistered,
		IOSNotificationResponseReasonBadDeviceToken: NotificationRequestErrorReasonIOSBadDeviceToken,
	}

	AndroidReason2NotificationErrorReason = map[string]NotificationRequestErrorReason{
		AndroidNotificationResponseReasonMismatchSenderId:    NotificationRequestErrorReasonAndroidMismatchSenderId,
		AndroidNotificationResponseReasonNotRegistered:       NotificationRequestErrorReasonAndroidNotRegistered,
		AndroidNotificationResponseReasonInvalidRegistration: NotificationRequestErrorReasonAndroidInvalidRegistration,
		AndroidNotificationResponseReasonInternalServerError: NotificationRequestErrorReasonAndroidInternalServerError,
	}

	ErrInvalidNotificationErrorReason = errors.New("invalid Notification Error Reason")
	ErrInvalidNotificationDeviceOS    = errors.New("invalid Notification Device OS")
)

func (e *DataNotificationError) Error() string {
	return fmt.Sprintf("failed sending data notification to %v tokens", len(e.NotificationErrors))
}

func (e *NotificationError) Error() string {
	return e.Err.Error()
}

func New(err, token string, deviceOS MobileOS, reason NotificationRequestErrorReason) (ne NotificationRequestError) {
	return NotificationRequestError{
		ErrMsg:   err,
		TokenStr: token,
		DeviceOS: deviceOS,
		Reason:   reason,
	}
}

func (e *NotificationRequestError) Error() string {
	return fmt.Sprintf("failed sending %s notification to %s: %s - %s", e.DeviceOS, e.TokenStr, e.Reason, e.ErrMsg)
}

func GetNotificationRequestErrorReason(deviceOS MobileOS, responseReason string) (NotificationRequestErrorReason, error) {
	var reasonMap map[string]NotificationRequestErrorReason

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
		return NotificationRequestErrorReason(responseReason), nil
	}
	return reason, nil
}
