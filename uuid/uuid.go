package uuid

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gofrs/uuid"
)

const (
	uuidFormat = "00000000-0000-0000-0000-000000000000"
)

var ErrInvalidSeqUUID = errors.New("invalid seq UUID")
var ErrInvalidUUID = errors.New("invalid UUID")

func NewUuidFromSeqID(ID uint) string {

	intID := int(ID)
	strID := strconv.Itoa(intID)

	lID := len(strID)

	lU := len(uuidFormat)

	runes := []rune(uuidFormat)
	substring := string(runes[0 : lU-lID])

	// "-" is not considered in this logic, and it will support 999,999,999,999 IDs
	return fmt.Sprintf("%s%s", substring, strID)

}

func NewSeqIDFromUuid(ID string) uint {

	runes := []rune(ID)
	substring := string(runes[24:])

	intID, _ := strconv.Atoi(substring)

	return uint(intID)

}

func NewUuid() string {

	newUuid, _ := uuid.NewV4()

	return newUuid.String()

}

// ValidateSeqUuid checks if the given uuid has '00000000-0000-0000-0000-' as its prefix, and its the other part is convertable into an int.
func ValidateSeqUuid(seqUUID string) error {
	uuidFormatRunes, uuidRune := []rune(uuidFormat), []rune(seqUUID)

	if len(uuidFormatRunes) != len(uuidRune) {
		return ErrInvalidSeqUUID
	}

	prefix := string(uuidFormatRunes[:24]) // 00000000-0000-0000-0000-
	uuidPrefix, uuidSeq := string(uuidRune[:24]), string(uuidRune[24:])

	if prefix != uuidPrefix {
		return ErrInvalidSeqUUID
	}

	_, err := strconv.Atoi(uuidSeq)
	if err != nil {
		return ErrInvalidSeqUUID
	}
	return nil
}

// ValidateUuid checks if the given uuid is in the format of 'ba0dd379-e482-4698-9a5d-2433b3c84a0a'
func ValidateUuid(stringUUID string) error {
	_, err := uuid.FromString(stringUUID)
	if err != nil {
		return ErrInvalidUUID
	}
	return nil
}
