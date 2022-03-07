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
func ValidateSeqUuid(uuid string) error {
	uuidFormatRunes, uuidRune := []rune(uuidFormat), []rune(uuid)

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
