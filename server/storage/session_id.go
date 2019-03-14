package storage

import (
	"fmt"
	"strconv"
	"strings"
)

// SessionID is a unique identifier for a measurement streaming session;
// sessions are ordered by this value
type SessionID = uint32

// SessionIDToString ...
func SessionIDToString(id SessionID) string {
	return fmt.Sprintf("%010d", id)
}

// SessionIDFromString ...
func SessionIDFromString(s string) (SessionID, error) {
	ix := strings.IndexFunc(s, func(r rune) bool { return r != '0' })
	if ix < 0 {
		return SessionID(0), nil
	}

	i, err := strconv.Atoi(s[ix:])
	if err != nil {
		return SessionID(0), err
	}

	return SessionID(uint32(i)), nil
}
