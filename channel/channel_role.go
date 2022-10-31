package channel

import "errors"

const (
	// valid channel roles
	ChannelRoleCareNav   = "carenav"
	ChannelRoleModerator = "moderator"

	ChannelRoleAutoResponder = "autoresponder"

	ChannelRoleMember = "member"

	ChannelRoleTriage = "triage" // not in use

	// a placeholder for channel roles, which is an invalid channel role
	ChannelRoleUnknown = "unknown"
)

var ErrInvalidChannelRole = errors.New("invalid channel role")

// GetStandardChannelRole converts a given string value for role into one of the predefined channel role constants.
func GetStandardChannelRole(role string) string {
	switch role {
	case ChannelRoleMember:
		return ChannelRoleMember
	case ChannelRoleCareNav:
		return ChannelRoleCareNav
	case ChannelRoleModerator:
		return ChannelRoleModerator
	case ChannelRoleTriage:
		return ChannelRoleTriage
	case ChannelRoleAutoResponder:
		return ChannelRoleAutoResponder
	default:
		return ChannelRoleUnknown
	}
}

// ValidateChannelRole checks if role is one of the predefined valid roles, if not it returns ErrInvalidChannelRole if the role is not valid
func ValidateChannelRole(role string) (string, error) {
	standardRole := GetStandardChannelRole(role)
	if standardRole == ChannelRoleUnknown {
		return "", ErrInvalidChannelRole
	}
	return standardRole, nil
}
