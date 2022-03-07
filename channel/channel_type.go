package channel

import (
	"errors"
	"fmt"
	"strings"
)

const ChannelTypeSurfix = "channel"
const ChannelTypeSeperator = "-"

const (
	// valid channel types
	ChannelTypeCareNav   = ChannelRoleCareNav + ChannelTypeSeperator + ChannelTypeSurfix
	ChannelTypeModerator = ChannelRoleModerator + ChannelTypeSeperator + ChannelTypeSurfix

	ChannelTypeMember = ChannelRoleMember + ChannelTypeSeperator + ChannelTypeSurfix // not in use
	ChannelTypeTriage = ChannelRoleTriage + ChannelTypeSeperator + ChannelTypeSurfix // not in use

	// a placeholder for channel types, which is an invalid channel type
	ChannelTypeUnknown = ChannelRoleUnknown + ChannelTypeSeperator + ChannelTypeSurfix
)

var ErrInvalidChannelType = errors.New("invalid channel type")

// GetStandardChannelType converts a given string value for channel type into one of the predefined channel type constants.
func GetStandardChannelType(channelType string) string {
	switch channelType {
	case ChannelTypeMember:
		return ChannelTypeMember
	case ChannelTypeCareNav:
		return ChannelTypeCareNav
	case ChannelTypeModerator:
		return ChannelTypeModerator
	case ChannelTypeTriage:
		return ChannelTypeTriage
	default:
		return ChannelTypeUnknown
	}
}

// NewChannelType returns the propery channel type with the given role string
func NewChannelType(role string) (string, error) {
	standardRole, err := ValidateChannelRole(role)
	if err != nil {
		return "", err
	}

	channelType := fmt.Sprintf("%s%s%s", standardRole, ChannelTypeSeperator, ChannelTypeSurfix)
	return channelType, nil
}

// ValidateChannelType validates the format and value of the given channel type string. It returns the corresponding channel type, the channel role. If the given channel type is invalid, it returns `ErrInvalidChannelType`.
func ValidateChannelType(channelType string) (string, string, error) {
	parts := strings.Split(channelType, ChannelTypeSeperator)
	if len(parts) != 2 {
		return "", "", ErrInvalidChannelType
	}

	if parts[1] != ChannelTypeSurfix {
		return "", "", ErrInvalidChannelType
	}

	role := GetStandardChannelRole(parts[0])
	if role == ChannelRoleUnknown {
		return "", "", ErrInvalidChannelType
	}
	return channelType, role, nil
}

// GetChannelRoleFromChannelType returns the corresponding channel role from the given channel type. It returns `ErrInvalidChannelType` if the format is invalid.
func GetChannelRoleFromChannelType(channelType string) (string, error) {
	_, channelRole, err := ValidateChannelType(channelType)
	if err != nil {
		return "", err
	}
	return channelRole, nil
}
