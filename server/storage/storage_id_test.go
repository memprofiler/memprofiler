package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionID_String(t *testing.T) {
	id := SessionID(0)
	assert.Equal(t, "0000000000", SessionIDToString(id))
	id = SessionID(1)
	assert.Equal(t, "0000000001", SessionIDToString(id))
	id = SessionID(100)
	assert.Equal(t, "0000000100", SessionIDToString(id))
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
