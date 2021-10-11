package resources

import (
	"io"
	"os"
	"strconv"
	"strings"
)

func readUint64FromFile(file *os.File) (uint64, error) {
	var (
		rawNumber []byte
		number    uint64
		err       error
	)
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	if rawNumber, err = io.ReadAll(file); err != nil {
		return 0, err
	}
	if number, err = strconv.ParseUint(strings.TrimSpace(string(rawNumber)), 10, 64); err != nil {
		return 0, err
	}
	return number, nil
}
