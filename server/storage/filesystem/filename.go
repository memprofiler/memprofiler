package filesystem

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

type measurementID uint32

func (mmID measurementID) String() string {
	return fmt.Sprintf("%010d", mmID)
}

func measurementIDFromString(s string) (measurementID, error) {
	ix := strings.IndexFunc(s, func(r rune) bool { return r != '0' })
	if ix < 0 {
		return measurementID(0), nil
	}

	i, err := strconv.Atoi(s[ix:])
	if err != nil {
		return measurementID(0), err
	}

	return measurementID(uint32(i)), nil
}

func makeFilename(subdir string, mmID measurementID) string {
	return filepath.Join(subdir, mmID.String())
}
