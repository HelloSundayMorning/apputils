package notification

import "fmt"

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
	IOSUnregistered   = NotificationErrorReason("Unregistered")
	IOSBadDeviceToken = NotificationErrorReason("BadDeviceToken")

	AndroidMismatchSenderId    = NotificationErrorReason("MismatchSenderId")
	AndroidNotRegistered       = NotificationErrorReason("NotRegistered")
	AndroidInvalidRegistration = NotificationErrorReason("InvalidRegistration")
	AndroidInternalServerError = NotificationErrorReason("InternalServerError")
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
