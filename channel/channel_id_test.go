package channel

import (
	"testing"

	"github.com/HelloSundayMorning/apputils/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewChannelID(t *testing.T) {

	testCases := []struct {
		role              string
		userID            string
		expectedChannelID string
		expectedErr       error
	}{
		{
			role:              "carenav",
			userID:            "00000000-0000-0000-0000-000000001234",
			expectedChannelID: "carenav-channel_00000000-0000-0000-0000-000000001234",
			expectedErr:       nil,
		}, {
			role:              "moderator",
			userID:            "00000000-0000-0000-0000-000000001234",
			expectedChannelID: "moderator-channel_00000000-0000-0000-0000-000000001234",
			expectedErr:       nil,
		}, {
			role:              "member",
			userID:            "00000000-0000-0000-0000-000000001234",
			expectedChannelID: "member-channel_00000000-0000-0000-0000-000000001234",
			expectedErr:       nil,
		}, {
			role:              "carenav",
			userID:            "1234",
			expectedChannelID: "",
			expectedErr:       uuid.ErrInvalidSeqUUID,
		},
	}

	for _, tc := range testCases {
		actualChannelID, actualErr := NewChannelID(tc.role, tc.userID)
		assert.Equal(t, tc.expectedChannelID, actualChannelID)
		assert.Equal(t, tc.expectedErr, actualErr)
	}
}

func TestGetChannelTypeFromChannelID(t *testing.T) {
	testCases := []struct {
		channelIDName       string
		expectedChannelType string
		expectedErr         error
	}{
		{
			channelIDName:       "carenav-channel_00000000-0000-0000-0000-000000001234",
			expectedChannelType: "carenav-channel",
			expectedErr:         nil,
		}, {
			channelIDName:       "moderator-channel_00000000-0000-0000-0000-000000001234",
			expectedChannelType: "moderator-channel",
			expectedErr:         nil,
		}, {
			channelIDName:       "member-channel_00000000-0000-0000-0000-000000001234",
			expectedChannelType: "member-channel",
			expectedErr:         nil,
		}, {
			channelIDName:       "moderator-channel_1234",
			expectedChannelType: "moderator-channel",
			expectedErr:         nil,
		}, {
			channelIDName:       "moderator_00000000-0000-0000-0000-000000001234",
			expectedChannelType: "",
			expectedErr:         ErrInvalidChannelType,
		}, {
			channelIDName:       "00000000-0000-0000-0000-000000001234",
			expectedChannelType: "",
			expectedErr:         ErrInvalidChannelID,
		},
	}

	for _, tc := range testCases {
		actualChannelType, actualErr := GetChannelTypeFromChannelID(tc.channelIDName)
		assert.Equal(t, tc.expectedChannelType, actualChannelType, "Input: %s", tc.channelIDName)
		assert.Equal(t, tc.expectedErr, actualErr, "Input: %s", tc.channelIDName)
	}
}

func TestGetUUIDFromChannelID(t *testing.T) {
	testCases := []struct {
		channelIDName string
		expectedUUID  string
		expectedErr   error
	}{
		{
			channelIDName: "carenav-channel_00000000-0000-0000-0000-000000001234",
			expectedUUID:  "00000000-0000-0000-0000-000000001234",
			expectedErr:   nil,
		}, {
			channelIDName: "moderator-channel_00000000-0000-0000-0000-000000001234",
			expectedUUID:  "00000000-0000-0000-0000-000000001234",
			expectedErr:   nil,
		}, {
			channelIDName: "member-channel_00000000-0000-0000-0000-000000001234",
			expectedUUID:  "00000000-0000-0000-0000-000000001234",
			expectedErr:   nil,
		}, {
			channelIDName: "moderator-channel_1234",
			expectedUUID:  "",
			expectedErr:   uuid.ErrInvalidSeqUUID,
		}, {
			channelIDName: "moderator_00000000-0000-0000-0000-000000001234",
			expectedUUID:  "00000000-0000-0000-0000-000000001234",
			expectedErr:   nil,
		}, {
			channelIDName: "00000000-0000-0000-0000-000000001234",
			expectedUUID:  "",
			expectedErr:   ErrInvalidChannelID,
		},
	}

	for _, tc := range testCases {
		actualUUID, actualErr := GetUUIDFromChannelID(tc.channelIDName)
		assert.Equal(t, tc.expectedUUID, actualUUID)
		assert.Equal(t, tc.expectedErr, actualErr)
	}
}

func TestGetChannelRoleFromChannelID(t *testing.T) {
	testCases := []struct {
		channelIDName       string
		expectedChannelRole string
		expectedErr         error
	}{
		{
			channelIDName:       "carenav-channel_00000000-0000-0000-0000-000000001234",
			expectedChannelRole: "carenav",
			expectedErr:         nil,
		}, {
			channelIDName:       "moderator-channel_00000000-0000-0000-0000-000000001234",
			expectedChannelRole: "moderator",
			expectedErr:         nil,
		}, {
			channelIDName:       "member-channel_00000000-0000-0000-0000-000000001234",
			expectedChannelRole: "member",
			expectedErr:         nil,
		}, {
			channelIDName:       "moderator-channel_1234",
			expectedChannelRole: "moderator",
			expectedErr:         nil,
		}, {
			channelIDName:       "moderator_00000000-0000-0000-0000-000000001234",
			expectedChannelRole: "",
			expectedErr:         ErrInvalidChannelType,
		}, {
			channelIDName:       "00000000-0000-0000-0000-000000001234",
			expectedChannelRole: "",
			expectedErr:         ErrInvalidChannelID,
		},
	}

	for _, tc := range testCases {
		actualChannelRole, actualErr := GetChannelRoleFromChannelID(tc.channelIDName)
		assert.Equal(t, tc.expectedChannelRole, actualChannelRole)
		assert.Equal(t, tc.expectedErr, actualErr)
	}
}

func TestGetChannelIDPartsFromChannelID(t *testing.T) {
	testCases := []struct {
		channelIDName       string
		expectedChannelRole string
		expectedUUID        string
		expectedErr         error
	}{
		{
			channelIDName:       "carenav-channel_00000000-0000-0000-0000-000000001234",
			expectedChannelRole: "carenav",
			expectedUUID:        "00000000-0000-0000-0000-000000001234",
			expectedErr:         nil,
		}, {
			channelIDName:       "moderator-channel_00000000-0000-0000-0000-000000001234",
			expectedChannelRole: "moderator",
			expectedUUID:        "00000000-0000-0000-0000-000000001234",
			expectedErr:         nil,
		}, {
			channelIDName:       "member-channel_00000000-0000-0000-0000-000000001234",
			expectedChannelRole: "member",
			expectedUUID:        "00000000-0000-0000-0000-000000001234",
			expectedErr:         nil,
		}, {
			channelIDName:       "moderator-channel_1234",
			expectedChannelRole: "",
			expectedUUID:        "",
			expectedErr:         uuid.ErrInvalidSeqUUID,
		}, {
			channelIDName:       "moderator_00000000-0000-0000-0000-000000001234",
			expectedChannelRole: "",
			expectedErr:         ErrInvalidChannelType,
		}, {
			channelIDName:       "00000000-0000-0000-0000-000000001234",
			expectedChannelRole: "",
			expectedErr:         ErrInvalidChannelID,
		},
	}

	for _, tc := range testCases {
		actualChannelRole, actualUUID, actualErr := GetChannelIDPartsFromChannelID(tc.channelIDName)
		assert.Equal(t, tc.expectedChannelRole, actualChannelRole, "Input: %s", tc.channelIDName)
		assert.Equal(t, tc.expectedUUID, actualUUID, "Input: %s", tc.channelIDName)
		assert.Equal(t, tc.expectedErr, actualErr, "Input: %s", tc.channelIDName)
	}
}
