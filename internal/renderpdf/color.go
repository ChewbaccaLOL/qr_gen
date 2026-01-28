package renderpdf

import (
	"errors"
	"fmt"
	"strings"
)

type hexColor struct {
	R int
	G int
	B int
}

func parseHexColor(value string) (hexColor, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return hexColor{}, errors.New("color is empty")
	}
	if trimmed == "transparent" {
		return hexColor{R: 0, G: 0, B: 0}, nil
	}
	if trimmed == "black" {
		return hexColor{R: 0, G: 0, B: 0}, nil
	}
	if trimmed == "white" {
		return hexColor{R: 255, G: 255, B: 255}, nil
	}
	if strings.HasPrefix(trimmed, "#") {
		hex := trimmed[1:]
		switch len(hex) {
		case 3:
			r, err := parseHexNibble(hex[0])
			if err != nil {
				return hexColor{}, err
			}
			g, err := parseHexNibble(hex[1])
			if err != nil {
				return hexColor{}, err
			}
			b, err := parseHexNibble(hex[2])
			if err != nil {
				return hexColor{}, err
			}
			return hexColor{R: int(r * 17), G: int(g * 17), B: int(b * 17)}, nil
		case 6:
			r, err := parseHexByte(hex[0:2])
			if err != nil {
				return hexColor{}, err
			}
			g, err := parseHexByte(hex[2:4])
			if err != nil {
				return hexColor{}, err
			}
			b, err := parseHexByte(hex[4:6])
			if err != nil {
				return hexColor{}, err
			}
			return hexColor{R: int(r), G: int(g), B: int(b)}, nil
		case 8:
			r, err := parseHexByte(hex[0:2])
			if err != nil {
				return hexColor{}, err
			}
			g, err := parseHexByte(hex[2:4])
			if err != nil {
				return hexColor{}, err
			}
			b, err := parseHexByte(hex[4:6])
			if err != nil {
				return hexColor{}, err
			}
			return hexColor{R: int(r), G: int(g), B: int(b)}, nil
		default:
			return hexColor{}, fmt.Errorf("unsupported hex length: %d", len(hex))
		}
	}
	return hexColor{}, fmt.Errorf("unsupported color format: %s", value)
}

func parseHexByte(value string) (uint8, error) {
	if len(value) != 2 {
		return 0, errors.New("invalid hex byte")
	}
	hi, err := parseHexNibble(value[0])
	if err != nil {
		return 0, err
	}
	lo, err := parseHexNibble(value[1])
	if err != nil {
		return 0, err
	}
	return hi*16 + lo, nil
}

func parseHexNibble(ch byte) (uint8, error) {
	switch {
	case ch >= '0' && ch <= '9':
		return uint8(ch - '0'), nil
	case ch >= 'a' && ch <= 'f':
		return uint8(ch-'a') + 10, nil
	case ch >= 'A' && ch <= 'F':
		return uint8(ch-'A') + 10, nil
	default:
		return 0, fmt.Errorf("invalid hex digit: %c", ch)
	}
}
