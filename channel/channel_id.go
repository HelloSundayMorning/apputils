package channel

import (
	"errors"
	"fmt"
	"strings"

	"github.com/HelloSundayMorning/apputils/uuid"
)

const ChannelIDSeperator = "_"

var ErrInvalidChannelID = errors.New("invalid channel ID")

// NewChannelID generates a channel ID from the given channel role and user ID. It returns `ErrInvalidChannelRole` or `ErrInvalidSeqUUID` if the given channel role or user ID is invalid.
func NewChannelID(role, userID string) (string, error) {

	channelType, err := NewChannelType(role)
	if err != nil {
		return "", err
	}

	err = uuid.ValidateSeqUuid(userID)
	if err != nil {
		return "", err
	}

	channelID := fmt.Sprintf("%s%s%s", channelType, ChannelIDSeperator, userID)
	return channelID, nil
}

// GetChannelTypeFromChannelID returns a predefined channel type from the given channel ID. It returns `ErrInvalidChannelID` or `ErrInvalidChannelType` if the channel ID or chanenl type is invalid.
func GetChannelTypeFromChannelID(channelID string) (string, error) {
	parts := strings.Split(channelID, ChannelIDSeperator)
	if len(parts) != 2 {
		return "", ErrInvalidChannelID
	}

	channelType, _, err := ValidateChannelType(parts[0])
	if err != nil {
		return "", err
	}
	return channelType, nil
}

// GetUUIDFromChannelID returns a user ID from the given channel ID. It returns `ErrInvalidChannelID` or `ErrInvalidSeqUUID` if the channel ID or the user ID is invalid.
func GetUUIDFromChannelID(channelID string) (string, error) {
	parts := strings.Split(channelID, ChannelIDSeperator)
	if len(parts) != 2 {
		return "", ErrInvalidChannelID
	}

	err := uuid.ValidateSeqUuid(parts[1])
	if err != nil {
		return "", err
	}
	return parts[1], nil
}

// GetChannelRoleFromChannelID returns a predefined channel role from the given channel ID. It returns `ErrInvalidChannelID` or `ErrInvalidChannelType` if the channel ID is invalid.
func GetChannelRoleFromChannelID(channelID string) (string, error) {
	parts := strings.Split(channelID, ChannelIDSeperator)
	if len(parts) != 2 {
		return "", ErrInvalidChannelID
	}

	channelRole, err := GetChannelRoleFromChannelType(parts[0])
	if err != nil {
		return "", err
	}
	return channelRole, nil
}

// GetChannelTypeAndUUIDFromChannelID returns the channel type and channel user ID. It returns `ErrInvalidChannelID` or `ErrInvalidSeqUUID` is the channel type or the user ID is invalid.
func GetChannelTypeAndUUIDFromChannelID(channelID string) (string, string, error) {
	parts := strings.Split(channelID, ChannelIDSeperator)
	if len(parts) != 2 {
		return "", "", ErrInvalidChannelID
	}

	channelType, _, err := ValidateChannelType(parts[0])
	if err != nil {
		return "", "", err
	}

	err = uuid.ValidateSeqUuid(parts[1])
	if err != nil {
		return "", "", err
	}

	return channelType, parts[1], nil
}

// GetChannelIDPartsFromChannelID splits channelID and returns the channel role and user ID.
func GetChannelIDPartsFromChannelID(channelID string) (string, string, error) {
	parts := strings.Split(channelID, ChannelIDSeperator)
	if len(parts) != 2 {
		return "", "", ErrInvalidChannelID
	}

	channelRole, err := GetChannelRoleFromChannelType(parts[0])
	if err != nil {
		return "", "", err
	}

	err = uuid.ValidateSeqUuid(parts[1])
	if err != nil {
		return "", "", err
	}

	return channelRole, parts[1], nil
}
