package uuid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUuidFromSeqID(t *testing.T) {

	ID := uint(32)

	assert.Equal(t, "00000000-0000-0000-0000-000000000032", NewUuidFromSeqID(ID))

	ID = uint(0)

	assert.Equal(t, "00000000-0000-0000-0000-000000000000", NewUuidFromSeqID(ID))

	ID = uint(100)

	assert.Equal(t, "00000000-0000-0000-0000-000000000100", NewUuidFromSeqID(ID))

	ID = uint(999999999999)

	assert.Equal(t, "00000000-0000-0000-0000-999999999999", NewUuidFromSeqID(ID))

	ID = uint(10000)

	assert.Equal(t, "00000000-0000-0000-0000-000000010000", NewUuidFromSeqID(ID))

}

func TestNewSeqIDFromUuid(t *testing.T) {

	ID := "00000000-0000-0000-0000-000000000032"

	assert.Equal(t, uint(32), NewSeqIDFromUuid(ID))

	ID = "00000000-0000-0000-0000-000000000000"

	assert.Equal(t, uint(0), NewSeqIDFromUuid(ID))

	ID = "00000000-0000-0000-0000-000000000100"

	assert.Equal(t, uint(100), NewSeqIDFromUuid(ID))

	ID = "00000000-0000-0000-0000-000000010000"

	assert.Equal(t, uint(10000), NewSeqIDFromUuid(ID))

	ID = "00000000-0000-0000-0000-999999999999"

	assert.Equal(t, uint(999999999999), NewSeqIDFromUuid(ID))
}

func TestValidateSeqUuid(t *testing.T) {

	ID := "00000000-0000-0000-0000-000000000032"

	assert.Nil(t, ValidateSeqUuid(ID))

	ID = "00000000-0000-0000-0000-000000001234"

	assert.Nil(t, ValidateSeqUuid(ID))

	ID = "00000000-0000-0000-0000-111111111234"

	assert.Nil(t, ValidateSeqUuid(ID))

	ID = "00000000-0000-0000-00009111111111234"

	assert.Equal(t, ErrInvalidSeqUUID, ValidateSeqUuid(ID))

	ID = "00000000-0000-0000-0000-11111111123"

	assert.Equal(t, ErrInvalidSeqUUID, ValidateSeqUuid(ID))

	ID = "0000000-0000-0000-0000-111111111234"

	assert.Equal(t, ErrInvalidSeqUUID, ValidateSeqUuid(ID))

	ID = "111111111234"

	assert.Equal(t, ErrInvalidSeqUUID, ValidateSeqUuid(ID))

	ID = "00000000-0000-0000-0000-a11111111234"

	assert.Equal(t, ErrInvalidSeqUUID, ValidateSeqUuid(ID))

	ID = "00000000-0000-0000-0000-e11111111234"

	assert.Equal(t, ErrInvalidSeqUUID, ValidateSeqUuid(ID))
}
