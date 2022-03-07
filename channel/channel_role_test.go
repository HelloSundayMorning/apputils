package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStandardChannelRole(t *testing.T) {

	testCases := []struct {
		roleName     string
		expectedRole string
	}{
		{
			roleName:     "",
			expectedRole: ChannelRoleUnknown,
		}, {
			roleName:     "member",
			expectedRole: ChannelRoleMember,
		}, {
			roleName:     "carenav",
			expectedRole: ChannelRoleCareNav,
		}, {
			roleName:     "moderator",
			expectedRole: ChannelRoleModerator,
		}, {
			roleName:     "triage",
			expectedRole: ChannelRoleTriage,
		}, {
			roleName:     "unknown",
			expectedRole: ChannelRoleUnknown,
		}, {
			roleName:     "hello",
			expectedRole: ChannelRoleUnknown,
		},
	}

	for _, tc := range testCases {
		actualRole := GetStandardChannelRole(tc.roleName)
		assert.Equal(t, tc.expectedRole, actualRole)
	}
}

func TestValidateChannelRole(t *testing.T) {

	testCases := []struct {
		role              string
		expectChannelRole string
		expectedError     error
	}{
		{
			role:              "",
			expectChannelRole: "",
			expectedError:     ErrInvalidChannelRole,
		}, {
			role:              "",
			expectChannelRole: "",
			expectedError:     ErrInvalidChannelRole,
		}, {
			role:              "moderator",
			expectChannelRole: ChannelRoleModerator,
			expectedError:     nil,
		}, {
			role:              "carenav",
			expectChannelRole: ChannelRoleCareNav,
			expectedError:     nil,
		},
	}

	for _, tc := range testCases {
		actualChannelRole, actualError := ValidateChannelRole(tc.role)
		assert.Equal(t, tc.expectChannelRole, actualChannelRole, "Input: %s", tc.role)
		assert.Equal(t, tc.expectedError, actualError, "Input: %s", tc.role)
	}
}
