package qr

import (
	"fmt"
	"strings"

	"rsc.io/qr"
)

func EncodeMatrix(data string, errorLevel string) ([][]bool, error) {
	level, err := parseLevel(errorLevel)
	if err != nil {
		return nil, err
	}
	code, err := qr.Encode(data, level)
	if err != nil {
		return nil, err
	}
	size := code.Size
	matrix := make([][]bool, size)
	for y := 0; y < size; y++ {
		row := make([]bool, size)
		for x := 0; x < size; x++ {
			row[x] = code.Black(x, y)
		}
		matrix[y] = row
	}
	return matrix, nil
}

func parseLevel(level string) (qr.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "l":
		return qr.L, nil
	case "m":
		return qr.M, nil
	case "q":
		return qr.Q, nil
	case "h":
		return qr.H, nil
	default:
		return qr.M, fmt.Errorf("invalid error correction level '%s'", level)
	}
}
