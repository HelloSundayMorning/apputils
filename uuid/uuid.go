package uuid

import (
	"fmt"
	"github.com/satori/go.uuid"
	"strconv"
)

const (
	uuidFormat = "00000000-0000-0000-0000-000000000000"
)

func NewUuidFromSeqID(ID uint) string {

	intID := int(ID)
	strID := strconv.Itoa(intID)

	lID := len(strID)

	lU := len(uuidFormat)

	runes := []rune(uuidFormat)
	substring := string(runes[0:lU-lID])

	return fmt.Sprintf("%s%s", substring, strID)

}

func NewSeqIDFromUuid(ID string) uint {

	runes := []rune(ID)
	substring := string(runes[24:])

	intID, _ := strconv.Atoi(substring)

	return uint(intID)

}

func NewUuid() string {

	uuid, _ := uuid.NewV4()

	return uuid.String()

}
