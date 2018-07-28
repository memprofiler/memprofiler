package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionID_String(t *testing.T) {
	var id SessionID
	id = 0
	assert.Equal(t, "0000000000", id.String())
	id = 1
	assert.Equal(t, "0000000001", id.String())
	id = 100
	assert.Equal(t, "0000000100", id.String())
}

func TestSessionID_FromString(t *testing.T) {

	var (
		id  SessionID
		err error
	)

	id, err = SessionIDFromString("0000000000")
	assert.NoError(t, err)
	assert.Equal(t, SessionID(0), id)

	id, err = SessionIDFromString("0000000001")
	assert.NoError(t, err)
	assert.Equal(t, SessionID(1), id)

	id, err = SessionIDFromString("0000000100")
	assert.NoError(t, err)
	assert.Equal(t, SessionID(100), id)
}
