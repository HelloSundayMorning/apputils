package uuid

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
