package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetChannelType(t *testing.T) {
	testCases := []struct {
		channelType         string
		expectedChannelType string
	}{
		{
			channelType:         "carenav-channel",
			expectedChannelType: ChannelTypeCareNav,
		}, {
			channelType:         "moderator-channel",
			expectedChannelType: ChannelTypeModerator,
		}, {
			channelType:         "hello-channel",
			expectedChannelType: ChannelTypeUnknown,
		}, {
			channelType:         "123-channel",
			expectedChannelType: ChannelTypeUnknown,
		}, {
			channelType:         "channel",
			expectedChannelType: ChannelTypeUnknown,
		}, {
			channelType:         "",
			expectedChannelType: ChannelTypeUnknown,
		},
	}

	for _, tc := range testCases {
		actualChannelType := GetStandardChannelType(tc.channelType)
		assert.Equal(t, tc.expectedChannelType, actualChannelType)
	}
}

func TestNewChannelType(t *testing.T) {

	testCases := []struct {
		role                string
		validRoles          []string
		expectedChannelType string
		expectedError       error
	}{
		{
			role:                "carenav",
			expectedChannelType: "carenav-channel",
			expectedError:       nil,
		}, {
			role:                "moderator",
			expectedChannelType: "moderator-channel",
			expectedError:       nil,
		}, {
			role:                "",
			expectedChannelType: "",
			expectedError:       ErrInvalidChannelRole,
		},
	}

	for _, tc := range testCases {
		actualChannelType, actualError := NewChannelType(tc.role)
		assert.Equal(t, tc.expectedChannelType, actualChannelType)
		assert.Equal(t, tc.expectedError, actualError)
	}
}

func TestValidateChannelType(t *testing.T) {

	testCases := []struct {
		channelType         string
		expectedChannelType string
		expectedChannelRole string
		expectedError       error
	}{
		{
			channelType:         "carenav-channel",
			expectedChannelType: "carenav-channel",
			expectedChannelRole: "carenav",
			expectedError:       nil,
		}, {
			channelType:         "moderator-channel",
			expectedChannelType: "moderator-channel",
			expectedChannelRole: "moderator",
			expectedError:       nil,
		}, {
			channelType:         "unknown-channel",
			expectedChannelType: "",
			expectedChannelRole: "",
			expectedError:       ErrInvalidChannelType,
		}, {
			channelType:         "moderator",
			expectedChannelType: "",
			expectedChannelRole: "",
			expectedError:       ErrInvalidChannelType,
		}, {
			channelType:         "",
			expectedChannelType: "",
			expectedChannelRole: "",
			expectedError:       ErrInvalidChannelType,
		},
	}

	for _, tc := range testCases {
		actualChannelType, actualChannelRole, actualError := ValidateChannelType(tc.channelType)
		assert.Equal(t, tc.expectedChannelType, actualChannelType)
		assert.Equal(t, tc.expectedChannelRole, actualChannelRole)
		assert.Equal(t, tc.expectedError, actualError)
	}
}

func TestGetChannelRoleFromChannelType(t *testing.T) {

	testCases := []struct {
		channelType         string
		expectedChannelRole string
		expectedError       error
	}{
		{
			channelType:         "carenav-channel",
			expectedChannelRole: "carenav",
			expectedError:       nil,
		}, {
			channelType:         "moderator-channel",
			expectedChannelRole: "moderator",
			expectedError:       nil,
		}, {
			channelType:         "moderator",
			expectedChannelRole: "",
			expectedError:       ErrInvalidChannelType,
		}, {
			channelType:         "",
			expectedChannelRole: "",
			expectedError:       ErrInvalidChannelType,
		},
	}

	for _, tc := range testCases {
		actualChannelRole, actualError := GetChannelRoleFromChannelType(tc.channelType)
		assert.Equal(t, tc.expectedChannelRole, actualChannelRole)
		assert.Equal(t, tc.expectedError, actualError)
	}
}
